/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testClientID = "spiffe://example.org/ns/prj-1/app1"
	testOtherID  = "spiffe://example.org/ns/prj-1/app2"
)

func TestClientTLSConfigWithoutAClientCertificate(t *testing.T) {
	setClientTLSEnv(t, "", "", "", "")

	cfg, err := clientTLSConfig()
	require.NoError(t, err)
	assert.False(t, cfg.InsecureSkipVerify, "without a client certificate the server is checked by host name")
	assert.Nil(t, cfg.GetClientCertificate)
}

func TestClientTLSConfigRefusesAnIncompleteSetup(t *testing.T) {
	ca := newTestCA(t)
	_, cert, key := ca.svid(t, testClientID)
	bundle := ca.bundleFile(t)

	tests := map[string]struct{ cert, key, anchors string }{
		"a certificate without its key": {cert: cert, anchors: bundle},
		"a key without its certificate": {key: key, anchors: bundle},
		"no trust anchors":              {cert: cert, key: key},
		"a missing trust anchors file":  {cert: cert, key: key, anchors: filepath.Join(t.TempDir(), "absent.crt")},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setClientTLSEnv(t, tc.cert, tc.key, tc.anchors, "")
			_, err := clientTLSConfig()
			assert.Error(t, err)
		})
	}
}

func TestClientTLSConfigRefusesAServerIDThatIsNotASPIFFEID(t *testing.T) {
	ca := newTestCA(t)
	_, cert, key := ca.svid(t, testClientID)
	bundle := ca.bundleFile(t)

	for _, id := range []string{"ola", "https://example.org/ns/prj-1/app1", "spiffe:///ns/prj-1/app1"} {
		t.Run(id, func(t *testing.T) {
			setClientTLSEnv(t, cert, key, bundle, id)
			_, err := clientTLSConfig()
			assert.ErrorContains(t, err, "is not a SPIFFE ID")
		})
	}
}

func TestClientTLSConfigPresentsTheClientCertificate(t *testing.T) {
	ca := newTestCA(t)
	der, cert, key := ca.svid(t, testClientID)
	setClientTLSEnv(t, cert, key, ca.bundleFile(t), "")

	cfg, err := clientTLSConfig()
	require.NoError(t, err)
	got, err := cfg.GetClientCertificate(nil)
	require.NoError(t, err)
	require.NotEmpty(t, got.Certificate)
	assert.Equal(t, der, got.Certificate[0])
}

func TestClientTLSConfigHoldsTheServerToASPIFFEID(t *testing.T) {
	ca := newTestCA(t)
	_, clientCert, clientKey := ca.svid(t, testClientID)
	sameID, _, _ := ca.svid(t, testClientID)
	otherID, _, _ := ca.svid(t, testOtherID)
	otherCA, _, _ := newTestCA(t).svid(t, testClientID)

	tests := map[string]struct {
		serverIDEnv string
		server      []byte
		ok          bool
	}{
		"the client's own ID":                  {server: sameID, ok: true},
		"another ID":                           {server: otherID},
		"the client's own ID from another CA":  {server: otherCA},
		"the ID DAPR_SERVER_SPIFFE_ID names":   {serverIDEnv: testOtherID, server: otherID, ok: true},
		"the client's own ID when another set": {serverIDEnv: testOtherID, server: sameID},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setClientTLSEnv(t, clientCert, clientKey, ca.bundleFile(t), tc.serverIDEnv)
			cfg, err := clientTLSConfig()
			require.NoError(t, err)

			err = cfg.VerifyPeerCertificate([][]byte{tc.server}, nil)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func setClientTLSEnv(t *testing.T, cert, key, anchors, serverID string) {
	t.Helper()
	t.Setenv(clientTLSCertFileEnvVarName, cert)
	t.Setenv(clientTLSKeyFileEnvVarName, key)
	t.Setenv(tlsTrustAnchorsFileEnvVarName, anchors)
	t.Setenv(serverSPIFFEIDEnvVarName, serverID)
}

type testCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func newTestCA(t *testing.T) testCA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return testCA{cert: cert, key: key}
}

// svid issues an SVID for id. It returns the certificate's DER, and the files
// it wrote the certificate and its PKCS#8 key to.
func (ca testCA) svid(t *testing.T, id string) ([]byte, string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	u, err := url.Parse(id)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		URIs:         []*url.URL{u},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.cert, &key.PublicKey, ca.key)
	require.NoError(t, err)
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	dir := t.TempDir()
	return der, writePEM(t, dir, "svid.crt", "CERTIFICATE", der), writePEM(t, dir, "svid.key", "PRIVATE KEY", keyDER)
}

func (ca testCA) bundleFile(t *testing.T) string {
	t.Helper()
	return writePEM(t, t.TempDir(), "bundle.crt", "CERTIFICATE", ca.cert.Raw)
}

func writePEM(t *testing.T, dir, name, blockType string, der []byte) string {
	t.Helper()
	file := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(file, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der}), 0o600))
	return file
}
