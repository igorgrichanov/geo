package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"geo/internal/app"
	"geo/internal/controller/http/v1/address"
	"geo/internal/infrastructure/responder"
	resp "geo/internal/lib/api/address/addressResponse"
	"geo/internal/service/geo"
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

func TestAddressGeocodeHandler(t *testing.T) {
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
		req         address.GeocodeRequest
		want        resp.Response
		respStatus  int
		useCaseMock *mocks.Geo
		mockError   bool
	}{
		{
			name: "success",
			req: address.GeocodeRequest{
				Lat: "55.8481373",
				Lng: "37.6414907",
			},
			want: resp.Response{
				Addresses: []*geo.Address{
					{
						City:   "Москва",
						Street: "Снежная",
						House:  "4",
						Lat:    "55.8481373",
						Lon:    "37.6414907",
					},
				},
			},
			respStatus:  http.StatusOK,
			useCaseMock: mocks.NewGeo(t),
			mockError:   false,
		},
		{
			name: "incorrect request",
			req: address.GeocodeRequest{
				Lat: "",
				Lng: "37.6414907",
			},
			want: resp.Response{
				Addresses: nil,
			},
			respStatus:  http.StatusBadRequest,
			useCaseMock: nil,
			mockError:   false,
		},
		{
			name: "use case error",
			req: address.GeocodeRequest{
				Lat: "55.8481373",
				Lng: "37.6414907",
			},
			want: resp.Response{
				Addresses: nil,
			},
			respStatus:  http.StatusInternalServerError,
			useCaseMock: mocks.NewGeo(t),
			mockError:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := address.New(log, requestIdKey, tt.useCaseMock, responseManager)
			handler := http.HandlerFunc(controller.Geocode)

			body, err := json.Marshal(tt.req)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/api/address/geocode", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "1")
			if tt.useCaseMock != nil {
				ctxMock := context.WithValue(ctx, requestIdKey, "1")
				if tt.mockError {
					tt.useCaseMock.On("Geocode", ctxMock, tt.req.Lat, tt.req.Lng).
						Return(nil, geo.ErrInternal).Once()
				} else {
					tt.useCaseMock.On("Geocode", ctxMock, tt.req.Lat, tt.req.Lng).
						Return(tt.want.Addresses, nil).Once()
				}
			}

			handler.ServeHTTP(rr, req.WithContext(ctx))

			require.Equal(t, tt.respStatus, rr.Code)

			if tt.respStatus == http.StatusOK {
				var res resp.Response
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &res))
				require.Equal(t, tt.want, res)
			}
			if tt.useCaseMock != nil {
				tt.useCaseMock.AssertExpectations(t)
			}
		})
	}
}
