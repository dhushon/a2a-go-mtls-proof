package auth

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
)

// GetCertificateThumbprint returns the SHA-256 thumbprint of a certificate,
// formatted as a URL-safe base64 string (RFC 8705).
func GetCertificateThumbprint(cert *x509.Certificate) string {
	fingerprint := sha256.Sum256(cert.Raw)
	return base64.RawURLEncoding.EncodeToString(fingerprint[:])
}

func GetServerTLSConfig() (*tls.Config, error) {
	// Load server cert and key
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %w", err)
	}

	// Load CA cert
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS config
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // Require and verify client certs
	}, nil
}

func GetClientTLSConfig() (*tls.Config, error) {
	// Load Agent's mTLS Credentials
	cert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a shared TLS config for the API call
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		// This is important for the client to verify the server's cert
		ServerName: "localhost",
	}, nil
}
