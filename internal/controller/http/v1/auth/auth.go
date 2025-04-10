package auth

import (
	"context"
	"errors"
	"geo/internal/infrastructure/responder"
	"geo/internal/lib/api/auth/request"
	"geo/internal/lib/api/auth/response"
	"geo/internal/lib/logger/sl"
	"geo/internal/service"
	"geo/internal/service/auth"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Auther interface {
	Login(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)
	Register(http.ResponseWriter, *http.Request)
}

type Auth struct {
	log          *slog.Logger
	requestIdKey string
	uc           service.Auth
	responder    responder.Responder
}

func New(log *slog.Logger, requestIdKey string, uc service.Auth, responder responder.Responder) *Auth {
	return &Auth{
		log:          log,
		requestIdKey: requestIdKey,
		uc:           uc,
		responder:    responder,
	}
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
}

func (e *LoginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

func SendToken(token string) render.Renderer {
	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}
}

// @Summary		Log in to the api
//
// @Tags			auth
//
// @Description	Get the Bearer token using your Login and Password. If the token's lifetime has expired, you need to log in again. If you don't have an account, see /register endpoint
// @Param			credentials	body		request.CredentialsRequest	true	"your credentials"
// @Success		200			{object}	LoginResponse
// @Failure		400			{object}	response.ErrResponse	"invalid login/password format"
// @Failure		401			{object}	response.ErrResponse	"Invalid username or password"
// @Failure		500			{object}	response.ErrResponse
// @Router			/login [post]
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	const op = "controller.auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	data := &request.CredentialsRequest{}
	if err := render.Bind(r, data); err != nil {
		log.Error("error decoding request", sl.Err(err))
		a.responder.ErrorBadRequest(w, err)
		return
	}
	log.Info("request received", slog.Any("data", data))

	ctx := context.WithValue(r.Context(), a.requestIdKey, middleware.GetReqID(r.Context()))
	token, err := a.uc.Login(ctx, data.Login, data.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		log.Error("error when logging in", sl.Err(err))
		a.responder.ErrorUnauthorized(w, err)
		//render.Render(w, r, response.ErrInvalidCredentials())
		return
	} else if err != nil {
		log.Error("error when logging in", sl.Err(err))
		a.responder.ErrorInternal(w, err)
		//render.Render(w, r, response.ErrInternal())
		return
	}
	log.Info("user logged in successfully", slog.Any("response", SendToken(token)))
	//render.Render(w, r, SendToken(token))
	a.responder.OutputJSON(w, SendToken(token))
}

// @Summary		Log out from the server
//
// @Tags			auth
//
// @Description	Log out and revoke the Bearer token
// @Success		204	"Logged out successfully"
//
// @Failure		401	"Unauthorized: Token missing or invalid"
// @Header			401	{string}	WWW-Authenticate	"Bearer"
//
// @Failure		500	{object}	response.ErrResponse
// @Security		ApiKeyAuth
// @Router			/logout [delete]
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.logout.New"
	log := a.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	log.Info("request received")

	t, claims, err := jwtauth.FromContext(r.Context())

	if err != nil {
		log.Error("malformed token passed through middleware")
		a.responder.ErrorInternal(w, err)
		//render.Render(w, r, response.ErrInternal())
		return
	}
	if t == nil {
		log.Error("token is nil")
		a.responder.ErrorInternal(w, errors.New("token is nil"))
		return
	}
	ctx := context.WithValue(r.Context(), a.requestIdKey, middleware.GetReqID(r.Context()))
	err = a.uc.Logout(ctx, claims)
	if err != nil {
		log.Error("logout error", sl.Err(err))
		a.responder.ErrorInternal(w, err)
		//render.Render(w, r, response.ErrInternal())
		return
	}
	log.Info("logged out successfully")
	render.Render(w, r, response.NoContent())
	return
}

// @Summary		Register on the server
// @Tags			auth
// @Description	Choose a login and set up a password
// @Param			credentials	body		request.CredentialsRequest	true	"your credentials"
// @Success		201			{object}	response.Response			"User registered successfully"
// @Failure		400			{object}	response.ErrResponse		"Invalid request parameters"
// @Failure		500			{object}	response.ErrResponse
// @Router			/register [post]
func (a *Auth) Register(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.register.New"
	log := a.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	data := &request.CredentialsRequest{}
	if err := render.Bind(r, data); err != nil {
		log.Error("error decoding request", sl.Err(err))
		//render.Render(w, r, response.ErrBadRequest(err.Error()))
		a.responder.ErrorBadRequest(w, err)
		return
	}
	log.Info("request received", slog.Any("data", data))

	ctx := context.WithValue(r.Context(), a.requestIdKey, middleware.GetReqID(r.Context()))
	err := a.uc.Register(ctx, data.Login, data.Password)
	if errors.Is(err, auth.ErrBadRequest) {
		log.Error("error registering user", sl.Err(err))
		//render.Render(w, r, response.ErrBadRequest(auth.ErrBadRequest.Error()))
		a.responder.ErrorBadRequest(w, auth.ErrBadRequest)
		return
	} else if err != nil {
		log.Error("error registering user", sl.Err(err))
		//render.Render(w, r, response.ErrInternal())
		a.responder.ErrorInternal(w, err)
		return
	}

	log.Info("user registered successfully",
		slog.Any("response", response.Created("User registered successfully")))
	//render.Render(w, r, response.Created("User registered successfully"))
	a.responder.Created(w, "User registered successfully")
}
