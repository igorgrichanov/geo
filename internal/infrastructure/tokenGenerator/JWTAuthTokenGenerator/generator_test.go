package JWTAuthTokenGenerator

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"strings"
	"testing"
	"time"
)

func TestManager_Generate(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil,
		jwt.WithAcceptableSkew(time.Second))
	m := New(tokenAuth, time.Second*30)

	type args struct {
		userLogin string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid token",
			args: args{
				userLogin: "user",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Generate(tt.args.userLogin)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(got, ".") {
				t.Errorf("Generate() does not contain .\n got = \n%s", got)
			}
		})
	}
}
