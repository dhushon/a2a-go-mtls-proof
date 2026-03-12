package auth

import (
	"context"
	"fmt"
	"time"
	"a2a-go-mtls-proof/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// OBOExchange defines the interface for performing OAuth 2.0 Token Exchange (RFC 8693).
type OBOExchange interface {
	ExchangeToken(ctx context.Context, subjectToken string, audience string) (string, error)
}

// MockOBOExchange is a mock implementation for local development and testing.
type MockOBOExchange struct {
	// Thumbprint to bind to the mock token
	Thumbprint string
}

func (m *MockOBOExchange) ExchangeToken(ctx context.Context, subjectToken string, audience string) (string, error) {
	logger.Info("Using Mock OBO Exchange", "audience", audience, "bound_to", m.Thumbprint)
	
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "mock-idp",
			Subject:   "mock-user",
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		Confirmation: struct {
			X5tS256 string `json:"x5t#S256"`
		}{
			X5tS256: m.Thumbprint,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	return token.SignedString(jwt.UnsafeAllowNoneSignatureType)
}

// OAuthExchange is the real implementation that calls a provider (e.g., Okta).
type OAuthExchange struct {
	Domain      string
	ClientID    string
	MTLSDomain  string
}

func (e *OAuthExchange) ExchangeToken(ctx context.Context, subjectToken string, audience string) (string, error) {
	return "", fmt.Errorf("OAuthExchange implementation pending: requires mTLS http.Client integration")
}

// NewOBOExchange returns the appropriate implementation based on the provided server type.
func NewOBOExchange(serverType, domain, clientID, mtlsDomain, thumbprint string) OBOExchange {
	switch serverType {
	case "okta", "generic":
		return &OAuthExchange{
			Domain:     domain,
			ClientID:   clientID,
			MTLSDomain: mtlsDomain,
		}
	default:
		return &MockOBOExchange{Thumbprint: thumbprint}
	}
}
