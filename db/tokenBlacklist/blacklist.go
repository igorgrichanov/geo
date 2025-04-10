package tokenBlacklist

import (
	"errors"
)

var (
	JTIAlreadyExists = errors.New("UUID already exists")
	Expired          = errors.New("token expired")
)
