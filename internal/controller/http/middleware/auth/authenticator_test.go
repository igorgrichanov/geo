package auth

import (
	"context"
	resp "geo/internal/lib/api/auth/response"
	"geo/internal/service/mocks"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func GenerateCorrectToken(t *testing.T, login, secret, jti string, skew time.Duration, exp time.Time) string {
	tokenAuth := jwtauth.New("HS256", []byte(secret), nil,
		jwt.WithAcceptableSkew(skew))
	_, tokenStr, err := tokenAuth.Encode(map[string]interface{}{
		"iss": "localhost:8080",
		"sub": login,
		"aud": "localhost:8080",
		"iat": time.Now().UTC(),
		"exp": exp,
		"jti": jti,
	})
	if err != nil {
		t.Errorf(err.Error())
		return ""
	}
	return tokenStr
}

func GenerateIncorrectToken(t *testing.T, login, secret, jti string, skew time.Duration) string {
	tokenAuth := jwtauth.New("HS256", []byte(secret), nil, jwt.WithAcceptableSkew(skew))
	_, token, err := tokenAuth.Encode(map[string]interface{}{
		"iss": "localhost:8080",
		"sub": login,
		"aud": "localhost:8080",
		"iat": time.Now().UTC(),
	})
	if err != nil {
		t.Errorf(err.Error())
		return ""
	}
	return token
}

func TestAuthenticator(t *testing.T) {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	login, secret, jti := gofakeit.Name(), gofakeit.Animal(), gofakeit.UUID()
	skew := time.Second * 5
	exp := time.Now().UTC()
	correctToken := GenerateCorrectToken(t, login, secret, jti, skew, exp)
	expiredToken := GenerateCorrectToken(t, login, secret, jti, skew, time.Now().UTC().Add(-10*time.Hour))

	tests := []struct {
		name        string
		token       string
		useCaseMock *mocks.Auth
		mockResp    bool
		respStatus  int
	}{
		{
			name:        "success",
			token:       correctToken,
			useCaseMock: mocks.NewAuth(t),
			mockResp:    false,
			respStatus:  http.StatusOK,
		},
		{
			name:        "nil token",
			token:       "",
			useCaseMock: nil,
			mockResp:    false,
			respStatus:  http.StatusUnauthorized,
		},
		{
			name:        "incorrect token",
			token:       GenerateIncorrectToken(t, login, secret, jti, skew),
			useCaseMock: nil,
			mockResp:    false,
			respStatus:  http.StatusUnauthorized,
		},
		{
			name:        "token expired",
			token:       expiredToken,
			useCaseMock: nil,
			mockResp:    false,
			respStatus:  http.StatusUnauthorized,
		},
		{
			name:        "token is in blacklist",
			token:       correctToken,
			useCaseMock: mocks.NewAuth(t),
			mockResp:    true,
			respStatus:  http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(jwtauth.Verifier(jwtauth.New("HS256", []byte(secret), nil, jwt.WithAcceptableSkew(skew))))
			authenticator := NewAuthenticator(log, tt.useCaseMock)
			router.Use(authenticator.Middleware())
			router.Post("/", func(w http.ResponseWriter, r *http.Request) {
				render.Render(w, r, &resp.Response{
					HTTPStatusCode: http.StatusOK,
				})
			})

			req, err := http.NewRequest(http.MethodPost, "/", nil)
			require.NoError(t, err)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "1")
			if tt.useCaseMock != nil {
				tt.useCaseMock.On("IsTokenRevoked", mock.Anything, jti).Return(tt.mockResp).Once()
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req.WithContext(ctx))

			require.Equal(t, tt.respStatus, rr.Code)
		})
	}
}
