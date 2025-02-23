package handlers

import "github.com/go-playground/validator/v10"

var (
	v = validator.New()
)

type userValidator struct {
	validator *validator.Validate
}

func (u *userValidator) Validate(i interface{}) error {
	return u.validator.Struct(i)
}
