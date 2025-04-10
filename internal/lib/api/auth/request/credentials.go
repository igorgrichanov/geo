package request

import (
	"fmt"
	"net/http"
)

type CredentialsRequest struct {
	Login    string `json:"login" example:"admin"`
	Password string `json:"password" example:"123456"`
} //@name CredentialsRequest

func (lr *CredentialsRequest) Bind(r *http.Request) error {
	if lr.Login == "" || lr.Password == "" {
		return fmt.Errorf("invalid request")
	}
	return nil
}
