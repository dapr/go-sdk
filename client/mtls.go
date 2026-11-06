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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
)

const (
	clientTLSCertFileEnvVarName   = "DAPR_CLIENT_TLS_CERT_FILE"
	clientTLSKeyFileEnvVarName    = "DAPR_CLIENT_TLS_KEY_FILE"
	tlsTrustAnchorsFileEnvVarName = "DAPR_TLS_TRUST_ANCHORS_FILE"
	serverSPIFFEIDEnvVarName      = "DAPR_SERVER_SPIFFE_ID"
)

// clientTLSConfig is the TLS config for a TLS endpoint. Without a client
// certificate it is the default: the server is checked against the system
// roots and the host name.
//
// With DAPR_CLIENT_TLS_CERT_FILE and DAPR_CLIENT_TLS_KEY_FILE set, the client
// presents that certificate, and the server is held to a SPIFFE identity
// instead of a host name: its chain must lead to DAPR_TLS_TRUST_ANCHORS_FILE,
// and its certificate must carry DAPR_SERVER_SPIFFE_ID, or by default the
// client certificate's own SPIFFE ID. That is the shape of a daprd reached
// over mTLS, whose SVID names no host. The files are read again on every
// handshake, so a rotated certificate or bundle is picked up without a
// restart.
func clientTLSConfig() (*tls.Config, error) {
	certFile := os.Getenv(clientTLSCertFileEnvVarName)
	keyFile := os.Getenv(clientTLSKeyFileEnvVarName)
	if certFile == "" && keyFile == "" {
		return new(tls.Config), nil
	}
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("%s and %s must be set together", clientTLSCertFileEnvVarName, clientTLSKeyFileEnvVarName)
	}
	anchorsFile := os.Getenv(tlsTrustAnchorsFileEnvVarName)
	if anchorsFile == "" {
		return nil, fmt.Errorf("%s is required with a client certificate", tlsTrustAnchorsFileEnvVarName)
	}

	// Load both once now, so a bad path or file fails here rather than on
	// the first call.
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("loading the client certificate: %w", err)
	}
	if _, err = loadTrustAnchors(anchorsFile); err != nil {
		return nil, err
	}

	serverID := os.Getenv(serverSPIFFEIDEnvVarName)
	if serverID != "" {
		u, parseErr := url.Parse(serverID)
		if parseErr != nil || u.Scheme != "spiffe" || u.Host == "" {
			return nil, fmt.Errorf("%s %q is not a SPIFFE ID, such as spiffe://<trust domain>/<path>", serverSPIFFEIDEnvVarName, serverID)
		}
	} else {
		leaf, parseErr := x509.ParseCertificate(cert.Certificate[0])
		if parseErr != nil {
			return nil, fmt.Errorf("parsing the client certificate: %w", parseErr)
		}
		if serverID, err = spiffeID(leaf); err != nil {
			return nil, fmt.Errorf("the client certificate gives no server SPIFFE ID to expect, set %s: %w", serverSPIFFEIDEnvVarName, err)
		}
	}

	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			c, loadErr := tls.LoadX509KeyPair(certFile, keyFile)
			if loadErr != nil {
				return nil, fmt.Errorf("loading the client certificate: %w", loadErr)
			}
			return &c, nil
		},
		// The server's certificate names a SPIFFE ID, not a host, so Go's
		// host name check is replaced by VerifyPeerCertificate.
		InsecureSkipVerify: true, //nolint:gosec // VerifyPeerCertificate checks the chain and the SPIFFE ID.
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			return verifyServer(rawCerts, anchorsFile, serverID)
		},
	}, nil
}

// verifyServer checks that the server's chain leads to the trust anchors and
// that its certificate carries serverID.
func verifyServer(rawCerts [][]byte, anchorsFile, serverID string) error {
	if len(rawCerts) == 0 {
		return errors.New("the server presented no certificate")
	}
	roots, err := loadTrustAnchors(anchorsFile)
	if err != nil {
		return err
	}
	certs := make([]*x509.Certificate, len(rawCerts))
	for i, raw := range rawCerts {
		if certs[i], err = x509.ParseCertificate(raw); err != nil {
			return fmt.Errorf("parsing the server certificate: %w", err)
		}
	}
	intermediates := x509.NewCertPool()
	for _, c := range certs[1:] {
		intermediates.AddCert(c)
	}
	if _, err = certs[0].Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		return fmt.Errorf("verifying the server certificate: %w", err)
	}
	id, err := spiffeID(certs[0])
	if err != nil {
		return fmt.Errorf("the server certificate: %w", err)
	}
	if id != serverID {
		return fmt.Errorf("the server is %s, expected %s", id, serverID)
	}
	return nil
}

func loadTrustAnchors(file string) (*x509.CertPool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading the trust anchors: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("no certificates in the trust anchors file %s", file)
	}
	return roots, nil
}

// spiffeID returns the SPIFFE ID in a certificate, which an SVID carries as
// its only URI SAN.
func spiffeID(cert *x509.Certificate) (string, error) {
	var ids []*url.URL
	for _, u := range cert.URIs {
		if u.Scheme == "spiffe" {
			ids = append(ids, u)
		}
	}
	if len(ids) != 1 {
		return "", fmt.Errorf("carries %d SPIFFE IDs, want 1", len(ids))
	}
	return ids[0].String(), nil
}
