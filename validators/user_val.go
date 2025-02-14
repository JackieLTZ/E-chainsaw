package validators

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

func (r *CreateUserRequest) Validate() error {
	err := validate.Struct(r)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name":
				errMsgs = append(errMsgs, "Name must be between 2 and 100 characters")
			case "Email":
				errMsgs = append(errMsgs, "Invalid email address")
			}
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, "; "))
	}
	return nil
}
