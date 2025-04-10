package response

import (
	"github.com/go-chi/render"
	"net/http"
)

type Response struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status,omitempty"`
}

func (e *Response) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	if e.HTTPStatusCode == 401 {
		w.Header().Set("WWW-Authenticate", `Bearer`) // закончить
	}
	return nil
}

func Created(status string) render.Renderer {
	return &Response{
		HTTPStatusCode: http.StatusCreated,
		StatusText:     status,
	}
}

func NoContent() render.Renderer {
	return &Response{
		HTTPStatusCode: http.StatusNoContent,
	}
}
