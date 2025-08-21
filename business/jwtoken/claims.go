package jwtoken

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	JobNumber string           `json:"sub,omitzero"`
	ExpiresAt *jwt.NumericDate `json:"exp,omitzero"`
	NotBefore *jwt.NumericDate `json:"nbf,omitzero"`
	IssuedAt  *jwt.NumericDate `json:"iat,omitzero"`
}

func (c *Claims) GetExpirationTime() (*jwt.NumericDate, error) {
	return c.ExpiresAt, nil
}

func (c *Claims) GetIssuedAt() (*jwt.NumericDate, error) {
	return c.IssuedAt, nil
}

func (c *Claims) GetNotBefore() (*jwt.NumericDate, error) {
	return c.NotBefore, nil
}

func (c *Claims) GetIssuer() (string, error) {
	return "goproxy", nil
}

func (c *Claims) GetSubject() (string, error) {
	return c.JobNumber, nil
}

func (c *Claims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
