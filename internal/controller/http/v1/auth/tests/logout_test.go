package tests

import (
	"context"
	"geo/internal/app"
	"geo/internal/controller/http/v1/auth"
	"geo/internal/infrastructure/responder"
	service "geo/internal/service/auth"
	"geo/internal/service/mocks"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ptflp/godecoder"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func GenerateCorrectToken(t *testing.T, login, secret, jti string, skew time.Duration, exp time.Time) jwt.Token {
	tokenAuth := jwtauth.New("HS256", []byte(secret), nil,
		jwt.WithAcceptableSkew(skew))
	token, _, err := tokenAuth.Encode(map[string]interface{}{
		"iss": "localhost:8080",
		"sub": login,
		"aud": "localhost:8080",
		"iat": time.Now().UTC(),
		"exp": exp,
		"jti": jti,
	})
	if err != nil {
		t.Errorf(err.Error())
		return nil
	}
	return token
}

func TestLogout(t *testing.T) {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	requestIdKey := app.RequestIdKey
	login, secret, jti := gofakeit.Name(), gofakeit.Animal(), gofakeit.UUID()
	skew := time.Second * 5
	exp := time.Now().UTC()
	decoder := godecoder.NewDecoder(jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		DisallowUnknownFields:  true,
	})
	responseManager := responder.NewResponder(decoder, log)

	tests := []struct {
		name             string
		respStatus       int
		token            jwt.Token
		useCaseMock      *mocks.Auth
		useCaseMockError error
	}{
		{
			name:             "success",
			respStatus:       http.StatusNoContent,
			token:            GenerateCorrectToken(t, login, secret, jti, skew, exp),
			useCaseMock:      mocks.NewAuth(t),
			useCaseMockError: nil,
		},
		{
			name:             "invalid token",
			respStatus:       http.StatusInternalServerError,
			token:            nil,
			useCaseMock:      nil,
			useCaseMockError: nil,
		},
		{
			name:             "internal server error",
			respStatus:       http.StatusInternalServerError,
			token:            GenerateCorrectToken(t, login, secret, jti, skew, exp),
			useCaseMock:      mocks.NewAuth(t),
			useCaseMockError: service.ErrInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := auth.New(log, requestIdKey, tt.useCaseMock, responseManager)
			handler := http.HandlerFunc(controller.Logout)

			req, err := http.NewRequest(http.MethodPost, "api/logout/", nil)
			require.NoError(t, err)
			ctx := jwtauth.NewContext(req.Context(), tt.token, nil)
			ctx = context.WithValue(ctx, middleware.RequestIDKey, "1")
			ctxMock := context.WithValue(ctx, requestIdKey, "1")

			if tt.useCaseMock != nil {
				_, claims, err := jwtauth.FromContext(ctx)
				require.NoError(t, err)
				tt.useCaseMock.On("Logout", ctxMock, claims).Return(tt.useCaseMockError).Once()
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req.WithContext(ctx))
			require.Equal(t, tt.respStatus, rr.Code)

			if tt.useCaseMock != nil {
				tt.useCaseMock.AssertExpectations(t)
			}
		})
	}
}
