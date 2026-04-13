package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateSignUpUser(body web_helpers.SignupBodyRequest) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := ValidateUsername(body.Username)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
	}
	err = ValidatePassword(body.Password)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
	}
	err = ValidateConfirmPassword(body.Password, body.ConfirmPassword)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
		validationErrors = append(validationErrors, web_helpers.NewFieldError("confirm_password", err.Error()))
	}
	err = ValidateEmail(body.Email)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
	}
	return validationErrors
}

func ValidateUpdateUser(body web_helpers.UpdateUserProfileRequest) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	if body.Username == nil && body.Email == nil && body.Password == nil && body.AvatarUrl == nil {
		return append(validationErrors, web_helpers.NewFieldError("request", "Нет полей для обновления"))
	}

	if body.Username != nil {
		err := ValidateUsername(*body.Username)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("username", err.Error()))
		}
	}

	if body.Email != nil {
		err := ValidateEmail(*body.Email)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("email", err.Error()))
		}
	}

	if body.Password != nil {
		err := ValidatePassword(*body.Password)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("password", err.Error()))
		}
	}
	return validationErrors
}
