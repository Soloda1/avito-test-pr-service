package utils

import "github.com/go-playground/validator/v10"

var (
	validatorInstance = validator.New()
)

func Validate(v any) error {
	return validatorInstance.Struct(v)
}
