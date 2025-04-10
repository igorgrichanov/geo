package inMemoryUserStorage

import (
	"errors"
	"fmt"
	"geo/db/userStorage"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type Storage struct {
	Users map[string]string
	mu    sync.RWMutex
}

func New() *Storage {
	return &Storage{
		Users: make(map[string]string, 100),
	}
}

func (r *Storage) Register(login, password string) error {
	r.mu.RLock()
	_, exists := r.Users[login]
	r.mu.RUnlock()
	if exists {
		return fmt.Errorf("login \"%s\": %w", login, userStorage.ErrAlreadyRegistered)
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("login \"%s\": %w", login, err)
	}

	r.mu.Lock()
	r.Users[login] = hashedPassword
	r.mu.Unlock()
	return nil
}

func (r *Storage) Login(login, password string) error {
	r.mu.RLock()
	hashedPassword, ok := r.Users[login]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("login \"%s\": %w", login, userStorage.ErrNotFound)
	}
	if err := checkPassword(password, hashedPassword); err != nil {
		return fmt.Errorf("login \"%s\": %w", login, err)
	}

	return nil
}

func checkPassword(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return userStorage.ErrIncorrectPassword
	}
	return err
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%w: %v", userStorage.ErrHashingPassword, err)
	}
	return string(hashedPassword), nil
}
