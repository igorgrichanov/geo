package http

import (
	"fmt"
	"geo/internal/config"
	"geo/internal/controller"
	"geo/internal/controller/http/middleware/auth"
	"geo/internal/controller/http/middleware/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
)

type AuthMiddleware struct {
	Authenticator *auth.Authenticator
	Ja            *jwtauth.JWTAuth
}

//	@Title			Geoservice API
//	@Version		1.0
//	@Description	Geoservice API allows users to search for addresses and geocode locations.
//	@Description	It supports authentication via JWT tokens and follows RESTful principles.

//	@Host		localhost:8080
//	@BasePath	/api
//	@Schemes	http
//	@Accept		json
//	@Produce	json

//	@securitydefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Specify the Bearer token in the format `Bearer <your_token>`

// @Tag.name			address
// @Tag.description	Get array of addresses

// @Tag.name			auth
// @Tag.description	Authorization and authentication
func NewRouter(log *slog.Logger, cfg *config.Config, controllers *controller.Controllers, am *AuthMiddleware) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(logger.New(log))

	router.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(am.Ja))
			r.Use(am.Authenticator.Middleware())
			r.Route("/address", func(r chi.Router) {
				r.Post("/search", controllers.Address.Search)
				r.Post("/geocode", controllers.Address.Geocode)
			})
			r.Delete("/logout", controllers.Auth.Logout)
		})
		r.Post("/login", controllers.Auth.Login)
		r.Post("/register", controllers.Auth.Register)
	})
	router.Get("/swagger/my.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/my.yaml")
	})

	host := cfg.Geoservice.Host
	if host == "" {
		host = "localhost"
	}
	swaggerUrl := fmt.Sprintf("http://%s%s/swagger/my.yaml", host, cfg.Geoservice.Port)
	router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(swaggerUrl)))

	return router
}
