package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateTransaction(body web_helpers.TransactionRequest, transactionTypes []string, categoryTypes []string, currencyCodes []string) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := ValidateTransactionTitle(body.Title)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", err.Error()))
	}
	err = ValidateTransactionDescription(body.Description)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", err.Error()))
	}
	err = ValidateTransactionValue(body.Value)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("value", err.Error()))
	}
	err = ValidateTransactionType(body.Type, transactionTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("type", err.Error()))
	}
	err = validateCategory(body.Category, categoryTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("category", err.Error()))
	}
	return validationErrors
}
