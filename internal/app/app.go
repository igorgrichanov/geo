package app

import (
	"context"
	"errors"
	"geo/db/tokenBlacklist/inMemoryTokenBlacklist"
	"geo/db/userStorage/inMemoryUserStorage"
	"geo/internal/config"
	"geo/internal/controller"
	httpController "geo/internal/controller/http"
	authMW "geo/internal/controller/http/middleware/auth"
	addressController "geo/internal/controller/http/v1/address"
	authController "geo/internal/controller/http/v1/auth"
	"geo/internal/infrastructure/geoProvider/dadata"
	"geo/internal/infrastructure/repository/token"
	"geo/internal/infrastructure/repository/user"
	"geo/internal/infrastructure/responder"
	"geo/internal/infrastructure/tokenGenerator/JWTAuthTokenGenerator"
	"geo/internal/lib/logger/sl"
	"geo/internal/service/auth"
	"geo/internal/service/geo"
	"github.com/go-chi/jwtauth/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ptflp/godecoder"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const RequestIdKey = "request_id"

func Run(cfg *config.Config) {
	log := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
	)

	// infrastructure
	geoProvider := dadata.NewGeoService(cfg.Dadata.ApiKey, cfg.Dadata.ApiSecret)
	ja := jwtauth.New("HS256", []byte(cfg.Token.Secret), nil,
		jwt.WithAcceptableSkew(cfg.Token.Skew))
	tokenGenerator := JWTAuthTokenGenerator.New(ja, cfg.Token.TTL)
	decoder := godecoder.NewDecoder(jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
		DisallowUnknownFields:  true,
	})
	responseManager := responder.NewResponder(decoder, log)

	// db
	tokenDB := inMemoryTokenBlacklist.NewBlacklist(cfg.Token.Skew)
	userDB := inMemoryUserStorage.New()

	// repository
	tokenRepo := token.New(tokenDB)
	userRepo := user.New(userDB)

	// service
	authService := auth.New(log, RequestIdKey, tokenRepo, tokenGenerator, userRepo)
	geoService := geo.New(log, RequestIdKey, geoProvider)

	// controller
	authCtrl := authController.New(log, RequestIdKey, authService, responseManager)
	addressCtrl := addressController.New(log, RequestIdKey, geoService, responseManager)
	ctrl := controller.New(authCtrl, addressCtrl)

	// router
	authenticator := authMW.NewAuthenticator(log, authService)
	router := httpController.NewRouter(log, cfg, ctrl, &httpController.AuthMiddleware{
		Authenticator: authenticator,
		Ja:            ja,
	})

	// server
	addr := cfg.Geoservice.Host + cfg.Geoservice.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("starting server at http://" + addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(err.Error())
		}
	}()

	<-done
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("error while shutting down server", sl.Err(err))
	}
	// close storage
	log.Info("shut down successfully")
}
