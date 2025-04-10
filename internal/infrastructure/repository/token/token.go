package token

import (
	"errors"
	"geo/db/tokenBlacklist"
	"time"
)

var (
	JTIAlreadyExists = errors.New("UUID already exists")
	Expired          = errors.New("token expired")
)

type Blacklist interface {
	Contains(string) bool
	Add(string, time.Time) error
}

type Repository struct {
	blacklist Blacklist
}

func New(u Blacklist) *Repository {
	return &Repository{u}
}

func (r *Repository) IsBlacklisted(jti string) bool {
	return r.blacklist.Contains(jti)
}

func (r *Repository) Add(jti string, exp time.Time) error {
	err := r.blacklist.Add(jti, exp)
	if errors.Is(err, tokenBlacklist.JTIAlreadyExists) {
		return JTIAlreadyExists
	} else if errors.Is(err, tokenBlacklist.Expired) {
		return Expired
	}
	return err
}
