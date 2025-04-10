package address

import (
	"context"
	"fmt"
	"geo/internal/infrastructure/responder"
	"geo/internal/lib/api/address/addressResponse"
	"geo/internal/lib/logger/sl"
	"geo/internal/service"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Addresser interface {
	Geocode(http.ResponseWriter, *http.Request)
	Search(http.ResponseWriter, *http.Request)
}

type Address struct {
	log          *slog.Logger
	requestIdKey string
	uc           service.Geo
	responder    responder.Responder
}

func New(log *slog.Logger, requestIdKey string, uc service.Geo, responder responder.Responder) *Address {
	return &Address{log: log, requestIdKey: requestIdKey, uc: uc, responder: responder}
}

type GeocodeRequest struct {
	Lat string `json:"lat" example:"55.8481373"`
	Lng string `json:"lng" example:"37.6414907"`
}

func (gr *GeocodeRequest) Bind(r *http.Request) error {
	if gr.Lat == "" || gr.Lng == "" {
		return fmt.Errorf("lat, lng cannot be empty")
	}
	return nil
}

// @Summary	Array of addresses located at specified coordinates
// @Tags		address
// @Param		coordinates	body		GeocodeRequest	true	"object coordinates"
// @Success	200			{object}	addressResponse.Response
// @Failure	400			{object}	response.ErrResponse	"invalid lat or lng format"
//
// @Failure	401			"Unauthorized: Token missing or invalid"
// @Header		401			{string}	WWW-Authenticate	"Bearer"
//
// @Failure	500			{object}	response.ErrResponse
// @Security	ApiKeyAuth
// @Router		/address/geocode [post]
func (a *Address) Geocode(w http.ResponseWriter, r *http.Request) {
	const op = "controller.address.Geocode"
	log := a.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	data := &GeocodeRequest{}
	if err := render.Bind(r, data); err != nil {
		log.Error("error decoding request", sl.Err(err))
		a.responder.ErrorBadRequest(w, err)
		return
	}
	log.Info("request received", slog.Any("data", data))

	ctx := context.WithValue(r.Context(), a.requestIdKey, middleware.GetReqID(r.Context()))
	addresses, err := a.uc.Geocode(ctx, data.Lat, data.Lng)
	if err != nil {
		log.Error("failed to get addresses using lat and lng", sl.Err(err))
		a.responder.ErrorInternal(w, err)
		return
	}

	log.Info("request executed", slog.Any("response", addressResponse.NewResponse(addresses)))
	a.responder.OutputJSON(w, addressResponse.NewResponse(addresses))
}

type SearchRequest struct {
	Query string `json:"query" example:"г Москва, ул Снежная"`
}

func (sr *SearchRequest) Bind(r *http.Request) error {
	if sr.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	return nil
}

// @Summary	Array of addresses located at specified location
// @Tags		address
// @Param		query	body		SearchRequest	true	"object location"
// @Success	200		{object}	addressResponse.Response
// @Failure	400		{object}	response.ErrResponse	"invalid query format"
//
// @Failure	401		"Unauthorized: Token missing or invalid"
// @Header		401		{string}	WWW-Authenticate	"Bearer"
//
// @Failure	500		{object}	response.ErrResponse
// @Security	ApiKeyAuth
// @Router		/address/search [post]
func (a *Address) Search(w http.ResponseWriter, r *http.Request) {
	const op = "controller.address.Search"
	log := a.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	data := &SearchRequest{}
	if err := render.Bind(r, data); err != nil {
		log.Error("error decoding request: ", sl.Err(err))
		a.responder.ErrorBadRequest(w, err)
		return
	}
	log.Info("request received", slog.Any("data", data))

	ctx := context.WithValue(r.Context(), a.requestIdKey, middleware.GetReqID(r.Context()))
	addresses, err := a.uc.Search(ctx, data.Query)
	if err != nil {
		log.Error("failed to get addresses using query \"%s\": ", data.Query, sl.Err(err))
		a.responder.ErrorInternal(w, err)
		return
	}
	log.Info("request executed", slog.Any("response", addressResponse.NewResponse(addresses)))

	a.responder.OutputJSON(w, addressResponse.NewResponse(addresses))
}
