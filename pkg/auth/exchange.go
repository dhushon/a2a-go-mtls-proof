package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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
	}
	claims.Confirmation.X5tS256 = m.Thumbprint

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	return token.SignedString(jwt.UnsafeAllowNoneSignatureType)
}

// OAuthExchange is the real implementation that calls a provider (e.g., Okta).
type OAuthExchange struct {
	Domain         string
	ClientID       string
	PrivateKeyPath string
	MTLSDomain     string
}

func (e *OAuthExchange) loadPrivateKey() (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(e.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first, then PKCS1
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

func (e *OAuthExchange) createClientAssertion() (string, error) {
	key, err := e.loadPrivateKey()
	if err != nil {
		return "", err
	}

	claims := jwt.RegisteredClaims{
		Issuer:    e.ClientID,
		Subject:   e.ClientID,
		Audience:  jwt.ClaimStrings{fmt.Sprintf("https://%s/oauth2/v1/token", e.Domain)},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "tgr8kWQ3CTDlFE8hCMe9EHfMVRE9BvWVuTWrPIpK9sA" // From user input

	return token.SignedString(key)
}

func (e *OAuthExchange) ExchangeToken(ctx context.Context, subjectToken string, audience string) (string, error) {
	if e.Domain == "" {
		return "", fmt.Errorf("OAUTH_DOMAIN is not configured")
	}

	// 1. Get mTLS Client
	client, err := GetMTLSClient("") 
	if err != nil {
		return "", fmt.Errorf("failed to create mTLS client: %w", err)
	}

	// 2. Create Client Assertion (private_key_jwt)
	assertion, err := e.createClientAssertion()
	if err != nil {
		return "", fmt.Errorf("failed to create client assertion: %w", err)
	}

	// 3. Build Token Exchange Request
	endpoint := fmt.Sprintf("https://%s/oauth2/v1/token", e.MTLSDomain)
	if e.MTLSDomain == "" {
		endpoint = fmt.Sprintf("https://%s/oauth2/v1/token", e.Domain)
	}

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	data.Set("client_id", e.ClientID)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", assertion)
	data.Set("subject_token", subjectToken)
	data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")
	data.Set("audience", audience)
	
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Info("Performing Okta Token Exchange (private_key_jwt)", "endpoint", endpoint, "audience", audience)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("okta returned error (%s): %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// NewOBOExchange returns the appropriate implementation based on the provided server type.
func NewOBOExchange(serverType, domain, clientID, privateKeyPath, mtlsDomain, thumbprint string) OBOExchange {
	switch serverType {
	case "okta", "generic":
		return &OAuthExchange{
			Domain:         domain,
			ClientID:       clientID,
			PrivateKeyPath: privateKeyPath,
			MTLSDomain:     mtlsDomain,
		}
	default:
		return &MockOBOExchange{Thumbprint: thumbprint}
	}
}
