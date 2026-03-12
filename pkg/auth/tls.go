package auth

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

// GetCertificateThumbprint returns the SHA-256 thumbprint of a certificate,
// formatted as a URL-safe base64 string (RFC 8705).
func GetCertificateThumbprint(cert *x509.Certificate) string {
	fingerprint := sha256.Sum256(cert.Raw)
	return base64.RawURLEncoding.EncodeToString(fingerprint[:])
}

// GetServerTLSConfig returns a config for responders.
func GetServerTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %w", err)
	}

	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}

// GetClientTLSConfig returns a config for requesters. 
func GetClientTLSConfig(serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	// For local mTLS (talking to our agents), we need our Test CA.
	// For external mTLS (talking to Okta), we need system CAs.
	var rootCAs *x509.CertPool
	
	if serverName == "localhost" || serverName == "127.0.0.1" {
		caCert, err := os.ReadFile("certs/ca.crt")
		if err == nil {
			rootCAs = x509.NewCertPool()
			rootCAs.AppendCertsFromPEM(caCert)
		}
	} else {
		// Use system CAs
		rootCAs, _ = x509.SystemCertPool()
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	}
	
	if serverName != "" {
		config.ServerName = serverName
	}

	return config, nil
}

// GetMTLSClient returns an http.Client configured with the agent's mTLS certificate.
func GetMTLSClient(serverName string) (*http.Client, error) {
	tlsConfig, err := GetClientTLSConfig(serverName)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}
