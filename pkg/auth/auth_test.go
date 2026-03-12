package auth

import (
	"crypto/x509"
	"testing"
)

func TestGetCertificateThumbprint(t *testing.T) {
	cert := &x509.Certificate{
		Raw: []byte("dummy-cert-data"),
	}
	
	thumbprint := GetCertificateThumbprint(cert)
	if thumbprint == "" {
		t.Error("expected non-empty thumbprint")
	}
	
	// Test consistency
	thumbprint2 := GetCertificateThumbprint(cert)
	if thumbprint != thumbprint2 {
		t.Error("thumbprint should be deterministic")
	}
}
