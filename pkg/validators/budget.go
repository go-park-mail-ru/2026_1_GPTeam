package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateBudget(body web_helpers.BudgetRequest, currencyCodes []string) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := validateBudgetTitle(body.Title)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", err.Error()))
	}
	err = validateBudgetDescription(body.Description)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", err.Error()))
	}
	err = validateCurrency(body.Currency, currencyCodes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("currency", err.Error()))
	}
	err = validateTargetBudget(body.Target)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("target", err.Error()))
	}
	err = validateActualBudget(body.Actual)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("actual", err.Error()))
	}
	err = validateBudgetStartDate(body.StartAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("start_at", err.Error()))
	}
	err = validateBudgetEndDate(body.StartAt, body.EndAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("end_at", err.Error()))
	}
	return validationErrors
}
