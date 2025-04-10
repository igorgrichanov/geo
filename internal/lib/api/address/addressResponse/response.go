package addressResponse

import (
	"geo/internal/service/geo"
	"net/http"
)

type Response struct {
	Addresses []*geo.Address `json:"addresses"`
} //@name AddressResponse

func NewResponse(addresses []*geo.Address) *Response {
	return &Response{
		Addresses: addresses,
	}
}

func (resp *Response) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
