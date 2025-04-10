package response

import (
	"github.com/go-chi/render"
	"net/http"
)

type ErrResponse struct {
	HTTPStatusCode   int    `json:"-"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	render.SetContentType(render.ContentTypeJSON)
	return nil
}

func ErrInvalidCredentials() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode:   http.StatusUnauthorized,
		Error:            "invalid_credentials",
		ErrorDescription: "Invalid username or password",
	}
}

func ErrBadRequest(errorText string) render.Renderer {
	return &ErrResponse{
		HTTPStatusCode:   http.StatusBadRequest,
		Error:            "invalid_request",
		ErrorDescription: errorText,
	}
}

func ErrRender() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode: http.StatusUnprocessableEntity,
		Error:          "Error rendering response.",
	}
}

func ErrInternal() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode:   http.StatusInternalServerError,
		Error:            "internal_error",
		ErrorDescription: "Something went wrong on the server",
	}
}

func ErrNotFound() render.Renderer {
	return &ErrResponse{
		HTTPStatusCode:   http.StatusNotFound,
		Error:            "Not found.",
		ErrorDescription: "The requested resource was not found.",
	}
}
