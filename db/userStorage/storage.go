package userStorage

import "errors"

var (
	ErrAlreadyRegistered = errors.New("user already registered in the system")
	ErrNotFound          = errors.New("user not found")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrHashingPassword   = errors.New("error hashing password")
)
