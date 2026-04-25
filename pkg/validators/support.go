package validators

import (
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/application/models"
	"github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"
)

func ValidateSupport(body web_helpers.SupportRequest, authUser models.UserModel) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := validateLength(body.Category, 1, 255)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("category", err.Error()))
	}
	err = validateLength(body.Message, 1, 255)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("message", err.Error()))
	}
	return validationErrors
}
