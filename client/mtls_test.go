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
	"context"
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

func TestClientTLSConfigIsNilWithNothingSet(t *testing.T) {
	setClientTLSEnv(t, "", "", "", "")

	cfg, err := clientTLSConfig()
	require.NoError(t, err)
	assert.Nil(t, cfg, "the endpoint alone decides whether to use TLS")
}

func TestTheTLSSettingsAreReadForAPlaintextEndpoint(t *testing.T) {
	setClientTLSEnv(t, "", "", "", "ola")

	_, err := NewClientWithAddressContext(context.Background(), "localhost:50001")
	assert.ErrorContains(t, err, "error configuring TLS")
}

func TestClientTLSConfigChecksTheHostNameWithoutAServerID(t *testing.T) {
	ca := newTestCA(t)
	_, cert, key := ca.svid(t, testClientID)
	caFile := ca.bundleFile(t)
	roots := x509.NewCertPool()
	roots.AddCert(ca.cert)

	tests := map[string]struct {
		cert, key, ca string
		roots         *x509.CertPool
		clientCert    bool
	}{
		"a CA alone":                    {ca: caFile, roots: roots},
		"a client certificate alone":    {cert: cert, key: key, clientCert: true},
		"a client certificate and a CA": {cert: cert, key: key, ca: caFile, roots: roots, clientCert: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setClientTLSEnv(t, tc.cert, tc.key, tc.ca, "")
			cfg, err := clientTLSConfig()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.False(t, cfg.InsecureSkipVerify, "Go checks the chain and the host name")
			assert.Nil(t, cfg.VerifyPeerCertificate)
			if tc.roots == nil {
				assert.Nil(t, cfg.RootCAs, "the system roots")
			} else {
				assert.True(t, tc.roots.Equal(cfg.RootCAs), "the CA set")
			}
			assert.Equal(t, tc.clientCert, cfg.GetClientCertificate != nil)
		})
	}
}

func TestClientTLSConfigRefusesAnIncompleteSetup(t *testing.T) {
	ca := newTestCA(t)
	_, cert, key := ca.svid(t, testClientID)
	caFile := ca.bundleFile(t)
	absent := filepath.Join(t.TempDir(), "absent.crt")

	tests := map[string]struct{ cert, key, ca, serverID string }{
		"a certificate without its key":      {cert: cert, ca: caFile},
		"a key without its certificate":      {key: key, ca: caFile},
		"a missing CA file":                  {cert: cert, key: key, ca: absent},
		"a server ID without a CA":           {cert: cert, key: key, serverID: testOtherID},
		"a server ID with a missing CA file": {cert: cert, key: key, ca: absent, serverID: testOtherID},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setClientTLSEnv(t, tc.cert, tc.key, tc.ca, tc.serverID)
			_, err := clientTLSConfig()
			assert.Error(t, err)
		})
	}
}

func TestClientTLSConfigRefusesAServerIDThatIsNotASPIFFEID(t *testing.T) {
	ca := newTestCA(t)
	_, cert, key := ca.svid(t, testClientID)
	caFile := ca.bundleFile(t)

	for _, id := range []string{"ola", "https://example.org/ns/prj-1/app1", "spiffe:///ns/prj-1/app1"} {
		t.Run(id, func(t *testing.T) {
			setClientTLSEnv(t, cert, key, caFile, id)
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

func TestClientTLSConfigHoldsTheServerToTheSPIFFEIDSet(t *testing.T) {
	ca := newTestCA(t)
	_, clientCert, clientKey := ca.svid(t, testClientID)
	wantID, _, _ := ca.svid(t, testOtherID)
	clientID, _, _ := ca.svid(t, testClientID)
	otherCA, _, _ := newTestCA(t).svid(t, testOtherID)

	tests := map[string]struct {
		server []byte
		ok     bool
	}{
		"the ID set":                  {server: wantID, ok: true},
		"the client's own ID":         {server: clientID},
		"the ID set, from another CA": {server: otherCA},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setClientTLSEnv(t, clientCert, clientKey, ca.bundleFile(t), testOtherID)
			cfg, err := clientTLSConfig()
			require.NoError(t, err)
			assert.True(t, cfg.InsecureSkipVerify, "the SPIFFE ID check replaces Go's host name check")

			err = cfg.VerifyPeerCertificate([][]byte{tc.server}, nil)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func setClientTLSEnv(t *testing.T, cert, key, ca, serverID string) {
	t.Helper()
	t.Setenv(grpcTLSCertPathEnvVarName, cert)
	t.Setenv(grpcTLSKeyPathEnvVarName, key)
	t.Setenv(grpcTLSCAPathEnvVarName, ca)
	t.Setenv(grpcTLSServerSPIFFEIDEnvVarName, serverID)
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
