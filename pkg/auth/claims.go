package auth

import "github.com/golang-jwt/jwt/v5"

// Claims defines the JWT structure including the 'cnf' (confirmation) claim.
// This is used for RFC 8705 certificate-bound access tokens.
type Claims struct {
	jwt.RegisteredClaims
	
	// RFC 8705 standard location
	Confirmation struct {
		X5tS256 string `json:"x5t#S256"` 
	} `json:"cnf"`

	// Okta-specific/Alternative location (since # is often forbidden in claim names)
	X5tS256TopLevel string `json:"x5t_S256"`
}

// GetThumbprint returns the certificate thumbprint from either the standard
// 'cnf' claim or the top-level 'x5t_S256' claim.
func (c *Claims) GetThumbprint() string {
	if c.Confirmation.X5tS256 != "" {
		return c.Confirmation.X5tS256
	}
	return c.X5tS256TopLevel
}
