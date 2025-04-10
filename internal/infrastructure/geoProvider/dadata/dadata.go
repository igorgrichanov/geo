package dadata

import (
	"context"
	"encoding/json"
	"fmt"
	provider "geo/internal/infrastructure/geoProvider"
	"geo/internal/service/geo"
	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
	"net/http"
	"net/url"
	"strings"
)

type GeoService struct {
	api       *suggest.Api
	apiKey    string
	secretKey string
}

func NewGeoService(apiKey, secretKey string) *GeoService {
	var err error
	endpointUrl, err := url.Parse("https://suggestions.dadata.ru/suggestions/api/4_1/rs/")
	if err != nil {
		return nil
	}

	creds := client.Credentials{
		ApiKeyValue:    apiKey,
		SecretKeyValue: secretKey,
	}

	api := suggest.Api{
		Client: client.NewClient(endpointUrl, client.WithCredentialProvider(&creds)),
	}

	return &GeoService{
		api:       &api,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

func (g *GeoService) AddressSearch(input string) ([]*geo.Address, error) {
	const op = "lib.geoProvider.dadata.AddressSearch"
	var res []*geo.Address
	rawRes, err := g.api.Address(context.Background(), &suggest.RequestParams{Query: input})
	if err != nil {
		return nil, provider.ErrUnavailable
	}

	for _, r := range rawRes {
		if r.Data.City == "" || r.Data.Street == "" {
			continue
		}
		res = append(res, &geo.Address{City: r.Data.City, Street: r.Data.Street, House: r.Data.House, Lat: r.Data.GeoLat, Lon: r.Data.GeoLon})
	}

	return res, nil
}

func (g *GeoService) AddressGeoCode(lat, lng string) ([]*geo.Address, error) {
	const op = "lib.geoProvider.dadata.AddressGeoCode"
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, lat, lng))
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address", data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", g.apiKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, provider.ErrUnavailable
	}
	var geoCode Response

	err = json.NewDecoder(resp.Body).Decode(&geoCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var res []*geo.Address
	for _, r := range geoCode.Suggestions {
		var address geo.Address
		address.City = string(r.Data.City)
		address.Street = string(r.Data.Street)
		address.House = r.Data.House
		address.Lat = r.Data.GeoLat
		address.Lon = r.Data.GeoLon

		res = append(res, &address)
	}

	return res, nil
}
