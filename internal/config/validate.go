package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func init() {
	const splitN = 2
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", splitN)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func validateTags(s interface{}) validator.ValidationErrors {
	if err := v.Struct(s); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			panic(fmt.Errorf("unexpected type from validation: %T", errs))
		}
		return errs
	}
	return nil
}

func Validate(config Config) []error {
	var res []error

	// run validations based on validator library
	errs := validateTags(config)
	for _, err := range errs {
		res = append(res, err)
	}

	// run hand-written validations
	validationFuncs := []func(Config) []error{placeholder}
	for _, f := range validationFuncs {
		errs := f(config)
		res = append(res, errs...)
	}

	return res
}

func placeholder(_ Config) []error {
	return nil
}
