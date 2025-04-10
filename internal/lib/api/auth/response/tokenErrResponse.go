package response

import (
	"fmt"
	"github.com/go-chi/render"
	"net/http"
)

// swaggerignore: true
type TokenErrResponse struct {
	HTTPStatusCode int    `json:"-"`
	Err            string `json:"-"`
	ErrDescription string `json:"-"`
}

func (re *TokenErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, re.HTTPStatusCode)
	if re.Err != "" && re.ErrDescription != "" {
		desc := fmt.Sprintf("Bearer, error=\"%s\", error_description=\"%s\"", re.Err, re.ErrDescription)
		w.Header().Set("WWW-Authenticate", desc)
	} else if re.Err != "" {
		desc := fmt.Sprintf("Bearer, error=\"%s\"", re.Err)
		w.Header().Set("WWW-Authenticate", desc)
	} else {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	return nil
}

func ErrTokenExpired() render.Renderer {
	return &TokenErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		Err:            "invalid_token",
		ErrDescription: "Token expired",
	}
}

func ErrTokenRevoked() render.Renderer {
	return &TokenErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		Err:            "invalid_token",
		ErrDescription: "Token has been revoked",
	}
}

func ErrTokenMalformed() render.Renderer {
	return &TokenErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
		Err:            "invalid_token",
		ErrDescription: "Token has been malformed",
	}
}

func ErrNoTokenProvided() render.Renderer {
	return &TokenErrResponse{
		HTTPStatusCode: http.StatusUnauthorized,
	}
}
