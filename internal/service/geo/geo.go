package geo

import (
	"context"
	"errors"
	"geo/internal/infrastructure/geoProvider"
	"geo/internal/lib/logger/sl"
	"log/slog"
)

var (
	ErrInternal = errors.New("internal server error")
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.3 --name=GeoProvider
type Provider interface {
	AddressSearch(input string) ([]*Address, error)
	AddressGeoCode(lat, lng string) ([]*Address, error)
}

type Address struct {
	City   string `json:"city"`
	Street string `json:"street"`
	House  string `json:"house"`
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
} //@name Address

type UseCase struct {
	log          *slog.Logger
	requestIdKey string
	provider     Provider
}

func New(log *slog.Logger, requestIDKey string, provider Provider) *UseCase {
	return &UseCase{
		log:          log,
		requestIdKey: requestIDKey,
		provider:     provider,
	}
}

func (s *UseCase) Geocode(ctx context.Context, lat, lng string) ([]*Address, error) {
	const op = "service.geo.Geocode"
	requestID := ctx.Value(s.requestIdKey).(string)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", requestID),
	)
	addresses, err := s.provider.AddressGeoCode(lat, lng)
	if errors.Is(err, geoProvider.ErrUnavailable) {
		log.Error("failed to get addresses", sl.Err(err))
		return nil, ErrInternal
	} else if err != nil {
		log.Error("failed to get addresses", sl.Err(err))
		return nil, ErrInternal
	}
	log.Info("addresses received")
	return addresses, nil
}

func (s *UseCase) Search(ctx context.Context, query string) ([]*Address, error) {
	const op = "service.geo.Search"
	requestID := ctx.Value(s.requestIdKey).(string)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", requestID),
	)
	addresses, err := s.provider.AddressSearch(query)
	if errors.Is(err, geoProvider.ErrUnavailable) {
		log.Error("failed to get addresses", sl.Err(err))
		return nil, ErrInternal
	} else if err != nil {
		log.Error("failed to get addresses", sl.Err(err))
		return nil, ErrInternal
	}
	log.Info("addresses received")
	return addresses, nil
}
