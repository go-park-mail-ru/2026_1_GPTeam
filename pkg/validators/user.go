package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateSignUpUser(body web_helpers.SignupBodyRequest) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := validateUsername(body.Username)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
	}
	err = validatePassword(body.Password)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
	}
	err = validateConfirmPassword(body.Password, body.ConfirmPassword)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("confirm_password", err.Error()))
	}
	err = validateEmail(body.Email)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
	}
	return validationErrors
}

func ValidateUpdateUser(body web_helpers.UpdateUserProfileRequest) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	if body.Username != nil {
		err := validateUsername(*body.Username)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
		}
	} else {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", "Поле не может быть пустым"))
	}
	if body.Email != nil {
		err := validateEmail(*body.Email)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
		}
	} else {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("email", "Поле не может быть пустым"))
	}
	if body.Password != nil {
		err := validatePassword(*body.Password)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
		}
	} else {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", "Поле не может быть пустым"))
	}
	return validationErrors
}
