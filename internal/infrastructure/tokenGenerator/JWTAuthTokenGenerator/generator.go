package JWTAuthTokenGenerator

import (
	"fmt"
	"geo/internal/infrastructure/tokenGenerator"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/jwtauth/v5"
	"time"
)

type JWTAuth struct {
	TokenAuth     *jwtauth.JWTAuth
	tokenLiveTime time.Duration
}

func New(tokenAuth *jwtauth.JWTAuth, tokenLiveTime time.Duration) *JWTAuth {
	return &JWTAuth{
		TokenAuth:     tokenAuth,
		tokenLiveTime: tokenLiveTime,
	}
}

func (m *JWTAuth) Generate(userLogin string) (string, error) {
	_, tokenString, err := m.TokenAuth.Encode(map[string]interface{}{
		"iss": "localhost:8080",
		"sub": userLogin,
		"aud": "localhost:8080",
		"iat": time.Now().UTC().Unix(),
		"exp": time.Now().UTC().Add(m.tokenLiveTime).Unix(),
		"jti": gofakeit.UUID(),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v, login \"%v\"", tokenGenerator.GenerationError, err, userLogin)
	}
	return tokenString, nil
}
