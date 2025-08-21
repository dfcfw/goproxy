package jwtoken

import (
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewIssue(secret []byte, log *slog.Logger) *Issue {
	if len(secret) == 0 {
		secret = make([]byte, 32)
		_, _ = rand.Read(secret)
	}

	return &Issue{
		secret: secret,
		log:    log,
	}
}

type Issue struct {
	secret []byte
	log    *slog.Logger
}

// Secret 获取 JWT 签名密钥。
func (iss *Issue) Secret() []byte {
	return iss.secret
}

// Sign 签发 JWT。
func (iss *Issue) Sign(jobNumber string, period time.Duration) (string, error) {
	now := time.Now()
	claim := &Claims{
		JobNumber: jobNumber,
		ExpiresAt: jwt.NewNumericDate(now.Add(period)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	return tk.SignedString(iss.secret)
}

func (iss *Issue) Valid(token string) (*Claims, error) {
	claim := new(Claims)
	_, err := jwt.ParseWithClaims(token, claim, iss.keyFunc)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (iss *Issue) keyFunc(*jwt.Token) (any, error) {
	return iss.secret, nil
}
