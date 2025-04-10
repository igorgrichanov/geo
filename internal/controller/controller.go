package controller

import (
	addressController "geo/internal/controller/http/v1/address"
	authController "geo/internal/controller/http/v1/auth"
)

type Controllers struct {
	Auth    authController.Auther
	Address addressController.Addresser
}

func New(auth authController.Auther, address addressController.Addresser) *Controllers {
	return &Controllers{
		Auth:    auth,
		Address: address,
	}
}
