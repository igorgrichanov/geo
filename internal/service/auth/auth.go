package auth

import (
	"context"
	"errors"
	"geo/internal/infrastructure/repository/token"
	"geo/internal/infrastructure/repository/user"
	"geo/internal/infrastructure/tokenGenerator"
	"geo/internal/lib/logger/sl"
	"log/slog"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternal           = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=Blacklister
type Blacklister interface {
	Add(jti string, exp time.Time) error
	IsBlacklisted(jti string) bool
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=TokenGenerator
type TokenGenerator interface {
	Generate(userLogin string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.0 --name=UserStorage
type UserStorage interface {
	LoginUser(string, string) error
	RegisterUser(string, string) error
}

type UseCase struct {
	log          *slog.Logger
	requestIdKey string
	bl           Blacklister
	tg           TokenGenerator
	us           UserStorage
}

func New(log *slog.Logger, requestIDKey string, bl Blacklister, tg TokenGenerator, us UserStorage) *UseCase {
	return &UseCase{
		log:          log,
		requestIdKey: requestIDKey,
		bl:           bl,
		tg:           tg,
		us:           us,
	}
}

func (s *UseCase) Login(ctx context.Context, login, password string) (string, error) {
	const op = "service.auth.Login"
	requestID := ctx.Value(s.requestIdKey).(string)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", requestID),
	)
	err := s.us.LoginUser(login, password)
	if errors.Is(err, user.ErrNotFound) {
		log.Error("user not found", sl.Err(err))
		return "", ErrInvalidCredentials
	} else if errors.Is(err, user.ErrIncorrectPassword) {
		log.Error("incorrect password", sl.Err(err))
		return "", ErrInvalidCredentials
	} else if err != nil {
		log.Error("failed to login", sl.Err(err))
		return "", ErrInternal
	}
	log.Info("user logged in successfully", sl.Info(login))

	t, err := s.tg.Generate(login)
	if errors.Is(err, tokenGenerator.GenerationError) {
		log.Error("error generating token", sl.Err(err))
		return "", ErrInternal
	} else if err != nil {
		log.Error("unable to generate token", sl.Err(err))
		return "", ErrInternal
	}
	log.Info("token generated", sl.Info(t), sl.Info(login))

	return t, nil
}

func (s *UseCase) Logout(ctx context.Context, claims map[string]interface{}) error {
	const op = "service.auth.Logout"
	requestID := ctx.Value(s.requestIdKey).(string)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", requestID),
	)
	jti, ok := claims["jti"].(string)
	if !ok {
		log.Error("malformed token passed through middleware")
		return ErrInternal
	}
	exp, ok := claims["exp"].(time.Time)
	if !ok {
		log.Error("malformed token passed through middleware")
		return ErrInternal
	}

	err := s.bl.Add(jti, exp)
	if errors.Is(err, token.JTIAlreadyExists) {
		log.Error("trying to add an existing jti into blacklist", sl.Err(err))
		return ErrInternal
	} else if errors.Is(err, token.Expired) {
		log.Error("trying to add an expired jti into blacklist", sl.Err(err))
		return ErrInternal
	} else if err != nil {
		log.Error("failed to add jti into blacklist", sl.Err(err))
		return ErrInternal
	}
	log.Info("jti invalidated", sl.Info(jti))
	return nil
}

func (s *UseCase) Register(ctx context.Context, login, password string) error {
	const op = "service.auth.RegisterUser"
	requestID := ctx.Value(s.requestIdKey).(string)
	log := s.log.With(
		slog.String("op", op),
		slog.String("request_id", requestID),
	)
	err := s.us.RegisterUser(login, password)
	if errors.Is(err, user.ErrAlreadyRegistered) {
		log.Error("user is already registered", sl.Err(err))
		return ErrBadRequest
	} else if errors.Is(err, user.ErrHashingPassword) {
		log.Error("error hashing password", sl.Err(err))
		return ErrInternal
	}

	log.Info("user registered successfully", sl.Info(login))
	return nil
}

func (s *UseCase) IsTokenRevoked(ctx context.Context, jti string) bool {
	return s.bl.IsBlacklisted(jti)
}
