package auth

import "github.com/golang-jwt/jwt/v5"

// Claims defines the JWT structure including the 'cnf' (confirmation) claim.
// This is used for RFC 8705 certificate-bound access tokens.
type Claims struct {
	jwt.RegisteredClaims
	Confirmation struct {
		X5tS256 string `json:"x5t#S256"` // SHA-256 thumbprint of the cert
	} `json:"cnf"`
}
