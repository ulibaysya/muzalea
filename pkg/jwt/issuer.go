package jwt

import "github.com/golang-jwt/jwt/v5"

// Issuer creates tokens and signs them.
type Issuer struct {
	method jwt.SigningMethod
	key    any
}

func NewIssuer(method jwt.SigningMethod, key any) Issuer {
	return Issuer{method, key}
}

func (i Issuer) Issue(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(i.method, claims).SignedString(i.key)
}
