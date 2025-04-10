package auth

import (
	"errors"
	resp "geo/internal/lib/api/auth/response"
	"geo/internal/lib/logger/sl"
	"geo/internal/service"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
)

type Authenticator struct {
	log *slog.Logger
	uc  service.Auth
}

func NewAuthenticator(log *slog.Logger, uc service.Auth) *Authenticator {
	return &Authenticator{
		log: log,
		uc:  uc,
	}
}

func (a *Authenticator) Middleware() func(http.Handler) http.Handler {
	const op = "controller.middleware.Authenticator.Middleware"
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.GetReqID(r.Context())
			log := a.log.With(
				slog.String("op", op),
				slog.String("request_id", requestID),
			)
			token, claims, err := jwtauth.FromContext(r.Context())
			if errors.Is(err, jwtauth.ErrExpired) {
				log.Error("token expired", sl.Err(err))
				render.Render(w, r, resp.ErrTokenExpired())
				return
			}
			if errors.Is(err, jwtauth.ErrNoTokenFound) {
				log.Error("token not found", sl.Err(err))
				render.Render(w, r, resp.ErrNoTokenProvided())
				return
			}
			if err != nil {
				log.Error("error getting token", sl.Err(err))
				render.Render(w, r, resp.ErrInternal())
			}
			if token == nil {
				log.Error("token is nil")
				render.Render(w, r, resp.ErrNoTokenProvided())
				return
			}
			jti, ok := claims["jti"].(string)
			if !ok {
				log.Error("jwt does not contain jti")
				render.Render(w, r, resp.ErrTokenMalformed())
				return
			}
			_, ok = claims["exp"].(time.Time)
			if !ok {
				log.Error("jwt does not contain exp")
				render.Render(w, r, resp.ErrTokenMalformed())
				return
			}
			if a.uc.IsTokenRevoked(r.Context(), jti) {
				log.Error("jwt is blacklisted")
				render.Render(w, r, resp.ErrTokenRevoked())
				return
			}
			log.Info("token accepted")
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
