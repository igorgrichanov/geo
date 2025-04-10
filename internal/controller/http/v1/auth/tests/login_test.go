package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"geo/internal/app"
	"geo/internal/controller/http/v1/auth"
	"geo/internal/infrastructure/responder"
	"geo/internal/lib/api/auth/request"
	service "geo/internal/service/auth"
	"geo/internal/service/mocks"
	"github.com/go-chi/chi/v5/middleware"
	jsoniter "github.com/json-iterator/go"
	"github.com/ptflp/godecoder"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	token := "123"
	requestIdKey := app.RequestIdKey
	decoder := godecoder.NewDecoder(jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		DisallowUnknownFields:  true,
	})
	responseManager := responder.NewResponder(decoder, log)

	tests := []struct {
		name             string
		req              request.CredentialsRequest
		wantResp         auth.LoginResponse
		respStatus       int
		useCaseMock      *mocks.Auth
		useCaseMockError error
	}{
		{
			name: "success",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			wantResp: auth.LoginResponse{
				AccessToken: token,
				TokenType:   "Bearer",
			},
			respStatus:       http.StatusOK,
			useCaseMock:      mocks.NewAuth(t),
			useCaseMockError: nil,
		},
		{
			name: "incorrect login or password",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "",
			},
			respStatus:       http.StatusBadRequest,
			useCaseMock:      nil,
			useCaseMockError: nil,
		},
		{
			name: "invalid credentials",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			respStatus:       http.StatusUnauthorized,
			useCaseMock:      mocks.NewAuth(t),
			useCaseMockError: service.ErrInvalidCredentials,
		},
		{
			name: "internal server error",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			respStatus:       http.StatusInternalServerError,
			useCaseMock:      mocks.NewAuth(t),
			useCaseMockError: service.ErrInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := auth.New(log, requestIdKey, tt.useCaseMock, responseManager)
			handler := http.HandlerFunc(controller.Login)

			body, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "api/login/", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "1")
			ctxMock := context.WithValue(ctx, requestIdKey, "1")
			if tt.useCaseMock != nil {
				tt.useCaseMock.On("Login", ctxMock, tt.req.Login, tt.req.Password).
					Return(token, tt.useCaseMockError).Once()
			}

			handler.ServeHTTP(rr, req.WithContext(ctx))
			require.Equal(t, tt.respStatus, rr.Code)

			// unmarshal ничего не загружает, если тип ответа не auth.LoginResponse
			var res auth.LoginResponse
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &res))
			require.Equal(t, tt.wantResp, res)

			if tt.useCaseMock != nil {
				tt.useCaseMock.AssertExpectations(t)
			}
		})
	}
}
