package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestNew(t *testing.T) {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)
	requestIdKey := app.RequestIdKey
	decoder := godecoder.NewDecoder(jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		DisallowUnknownFields:  true,
	})
	responseManager := responder.NewResponder(decoder, log)

	tests := []struct {
		name        string
		req         request.CredentialsRequest
		respStatus  int
		useCaseMock *mocks.Auth
		mockError   error
	}{
		{
			name: "success",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			respStatus:  http.StatusCreated,
			useCaseMock: mocks.NewAuth(t),
			mockError:   nil,
		},
		{
			name: "incorrect login or password format",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "",
			},
			respStatus:  http.StatusBadRequest,
			useCaseMock: nil,
			mockError:   nil,
		},
		{
			name: "incorrect login or password",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			respStatus:  http.StatusBadRequest,
			useCaseMock: mocks.NewAuth(t),
			mockError:   service.ErrBadRequest,
		},
		{
			name: "internal error",
			req: request.CredentialsRequest{
				Login:    "user",
				Password: "password",
			},
			respStatus:  http.StatusInternalServerError,
			useCaseMock: mocks.NewAuth(t),
			mockError:   errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := auth.New(log, requestIdKey, tt.useCaseMock, responseManager)
			handler := http.HandlerFunc(controller.Register)
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "api/register/", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "1")
			ctxMock := context.WithValue(ctx, requestIdKey, "1")
			if tt.useCaseMock != nil {
				tt.useCaseMock.On("Register", ctxMock, tt.req.Login, tt.req.Password).
					Return(tt.mockError).Once()
			}
			handler.ServeHTTP(rr, req.WithContext(ctx))

			require.Equal(t, tt.respStatus, rr.Code)

			if tt.useCaseMock != nil {
				tt.useCaseMock.AssertExpectations(t)
			}
		})
	}
}
