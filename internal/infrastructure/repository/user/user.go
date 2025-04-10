package user

import (
	"errors"
	"geo/db/userStorage"
)

var (
	ErrAlreadyRegistered = errors.New("user already registered in the system")
	ErrNotFound          = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrHashingPassword   = errors.New("error hashing password")
)

type Storage interface {
	Register(login, password string) error
	Login(login, password string) error
}

type Repository struct {
	storage Storage
}

func New(u Storage) *Repository {
	return &Repository{u}
}

func (ur *Repository) RegisterUser(login, password string) error {
	err := ur.storage.Register(login, password)
	if errors.Is(err, userStorage.ErrAlreadyRegistered) {
		return ErrAlreadyRegistered
	} else if errors.Is(err, userStorage.ErrHashingPassword) {
		return ErrHashingPassword
	}
	return err
}

func (ur *Repository) LoginUser(login, password string) error {
	err := ur.storage.Login(login, password)
	if errors.Is(err, userStorage.ErrIncorrectPassword) {
		return ErrIncorrectPassword
	} else if errors.Is(err, userStorage.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
