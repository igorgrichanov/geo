package service

import (
	"context"
	"geo/internal/service/geo"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=Auth
type Auth interface {
	Register(ctx context.Context, login, password string) error
	Logout(ctx context.Context, claims map[string]interface{}) error
	Login(ctx context.Context, login, password string) (string, error)
	IsTokenRevoked(ctx context.Context, jti string) bool
}

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=Geo
type Geo interface {
	Geocode(ctx context.Context, lat, lng string) ([]*geo.Address, error)
	Search(ctx context.Context, query string) ([]*geo.Address, error)
}
