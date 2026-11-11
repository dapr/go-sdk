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
	grpcTLSCertPathEnvVarName       = "DAPR_GRPC_TLS_CERT_PATH"
	grpcTLSKeyPathEnvVarName        = "DAPR_GRPC_TLS_KEY_PATH"
	grpcTLSCAPathEnvVarName         = "DAPR_GRPC_TLS_CA_PATH"
	grpcTLSServerSPIFFEIDEnvVarName = "DAPR_GRPC_TLS_SERVER_SPIFFE_ID"
)

// clientTLSConfig reads the TLS settings of the gRPC connection, which mean
// what they do in the Java SDK:
//
//   - DAPR_GRPC_TLS_CERT_PATH and DAPR_GRPC_TLS_KEY_PATH, set together, are
//     the certificate the client presents.
//   - DAPR_GRPC_TLS_CA_PATH is the CA the server's certificate must chain to,
//     in place of the system roots.
//
// The server's certificate must also name the host dialed, or the authority
// when one is set. DAPR_GRPC_TLS_SERVER_SPIFFE_ID replaces that check with a
// SPIFFE ID one, for a server whose certificate names no host, such as a
// daprd reached over mTLS: the certificate must carry that ID, and chain to
// DAPR_GRPC_TLS_CA_PATH, which is then required.
//
// The client certificate, and the CA when a SPIFFE ID is checked, are read
// again on every handshake, so a rotated one is picked up without a restart.
// It returns nil when none of the settings is set.
func clientTLSConfig() (*tls.Config, error) {
	certFile := os.Getenv(grpcTLSCertPathEnvVarName)
	keyFile := os.Getenv(grpcTLSKeyPathEnvVarName)
	caFile := os.Getenv(grpcTLSCAPathEnvVarName)
	serverID := os.Getenv(grpcTLSServerSPIFFEIDEnvVarName)
	if certFile == "" && keyFile == "" && caFile == "" && serverID == "" {
		return nil, nil
	}

	cfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			return nil, fmt.Errorf("%s and %s must be set together", grpcTLSCertPathEnvVarName, grpcTLSKeyPathEnvVarName)
		}
		// Load it once now, so a bad path or file fails here rather than on
		// the first call.
		if _, err := tls.LoadX509KeyPair(certFile, keyFile); err != nil {
			return nil, fmt.Errorf("loading the client certificate: %w", err)
		}
		cfg.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cert, loadErr := tls.LoadX509KeyPair(certFile, keyFile)
			if loadErr != nil {
				return nil, fmt.Errorf("loading the client certificate: %w", loadErr)
			}
			return &cert, nil
		}
	}

	switch {
	case serverID != "":
		u, parseErr := url.Parse(serverID)
		if parseErr != nil || u.Scheme != "spiffe" || u.Host == "" {
			return nil, fmt.Errorf("%s %q is not a SPIFFE ID, such as spiffe://<trust domain>/<path>", grpcTLSServerSPIFFEIDEnvVarName, serverID)
		}
		if caFile == "" {
			return nil, fmt.Errorf("%s requires %s", grpcTLSServerSPIFFEIDEnvVarName, grpcTLSCAPathEnvVarName)
		}
		if _, err := loadCA(caFile); err != nil {
			return nil, err
		}
		// The server's certificate names a SPIFFE ID, not a host, so Go's
		// host name check is replaced by VerifyPeerCertificate.
		cfg.InsecureSkipVerify = true //nolint:gosec // VerifyPeerCertificate checks the chain and the SPIFFE ID.
		cfg.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			return verifyServer(rawCerts, caFile, serverID)
		}
	case caFile != "":
		roots, err := loadCA(caFile)
		if err != nil {
			return nil, err
		}
		cfg.RootCAs = roots
	}

	return cfg, nil
}

// verifyServer checks that the server's chain leads to the CA and that its
// certificate carries serverID.
func verifyServer(rawCerts [][]byte, caFile, serverID string) error {
	if len(rawCerts) == 0 {
		return errors.New("the server presented no certificate")
	}
	roots, err := loadCA(caFile)
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

func loadCA(file string) (*x509.CertPool, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading the CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("no certificates in the CA file %s", file)
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
