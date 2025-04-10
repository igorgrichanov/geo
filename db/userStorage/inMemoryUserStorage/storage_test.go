package inMemoryUserStorage

import (
	"errors"
	"geo/db/userStorage"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestUserInMemoryRegistry_RegisterUser(t *testing.T) {
	s := New()
	pwdToHash := "test"
	pwd1, err := bcrypt.GenerateFromPassword([]byte(pwdToHash), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		login    string
		password string
	}
	tests := []struct {
		name       string
		args       args
		wantHashed string
		wantErr    error
	}{
		{
			name: "success",
			args: args{
				login:    "test",
				password: pwdToHash,
			},
			wantHashed: string(pwd1),
			wantErr:    nil,
		},
		{
			name: "existing user",
			args: args{
				login:    "test",
				password: pwdToHash,
			},
			wantHashed: string(pwd1),
			wantErr:    userStorage.ErrAlreadyRegistered,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.Register(tt.args.login, tt.args.password); !errors.Is(err, tt.wantErr) {
				t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			p, ok := s.Users[tt.args.login]
			if !ok {
				t.Errorf("user %v does not exist", tt.args.login)
			}
			if p == tt.wantHashed {
				t.Errorf("two equal passwords cannot have equal hashes: %v and %v",
					s.Users[tt.args.login], tt.wantHashed)
			}
		})
	}
}

func Test_hashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				password: "test",
			},
			wantErr: nil,
		},
		{
			name: "empty",
			args: args{
				password: "",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hashPassword(tt.args.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("hashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == tt.args.password {
				t.Errorf("passwords cannot be equal")
			}
		})
	}
}

func Test_checkPassword(t *testing.T) {
	pwd1, err := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	pwd2, err := bcrypt.GenerateFromPassword([]byte(""), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		password       string
		hashedPassword string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				password:       "test",
				hashedPassword: string(pwd1),
			},
			wantErr: nil,
		},
		{
			name: "empty",
			args: args{
				password:       "",
				hashedPassword: string(pwd2),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkPassword(tt.args.password, tt.args.hashedPassword); !errors.Is(err, tt.wantErr) {
				t.Errorf("checkPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserInMemoryRegistry_LoginUser(t *testing.T) {
	s := New()
	err := s.Register("test", "test")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		login    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				login:    "test",
				password: "test",
			},
			wantErr: nil,
		},
		{
			name: "failure",
			args: args{
				login:    "test",
				password: "wrong",
			},
			wantErr: userStorage.ErrIncorrectPassword,
		},
		{
			name: "unknown user",
			args: args{
				login:    "unknown",
				password: "test",
			},
			wantErr: userStorage.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Login(tt.args.login, tt.args.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
