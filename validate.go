package websocket

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
	"strings"
)

var (
	validate *validator.Validate
)

// init creates a validator instance and initializes the translator
func init() {
	validate = validator.New()
	// register custom validators
	err := validate.RegisterValidation("phone", validatePhoneNumber)
	if err != nil {
		return
	}
}

// ValidateStructWithOutCtx check struct
func ValidateStructWithOutCtx(s interface{}) (err error) {
	err = validate.Struct(s)
	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return
		}
		for _, e := range err.(validator.ValidationErrors) {
			return errors.New(
				fmt.Sprintf("params `%s` verification failed, errors: %s %s ",
					strings.ToLower(e.StructNamespace()), e.Tag(), e.Param()))
		}
	}
	return
}

func validatePhoneNumber(fl validator.FieldLevel) bool {
	phoneNumber := fl.Field().String()
	// Regular expression matching mobile phone number in Chinese Mainland.
	regex := `^1[3-9]\d{9}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(phoneNumber)
}
