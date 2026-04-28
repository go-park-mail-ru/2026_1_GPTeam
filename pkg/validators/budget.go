package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateBudget(body web_helpers.BudgetRequest, currencyCodes []string, allowedCategories []string) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := ValidateBudgetTitle(body.Title)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", err.Error()))
	}
	err = ValidateBudgetDescription(body.Description)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", err.Error()))
	}
	err = ValidateCurrency(body.Currency, currencyCodes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("currency", err.Error()))
	}
	err = ValidateTargetBudget(body.Target)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("target", err.Error()))
	}
	err = ValidateActualBudget(body.Actual)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("actual", err.Error()))
	}
	err = ValidateBudgetStartDate(body.StartAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("start_at", err.Error()))
	}
	err = ValidateBudgetEndDate(body.StartAt, body.EndAt)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("end_at", err.Error()))
	}
	if len(body.Category) == 0 {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("category", "Не выбрано категорий"))
	}
	for _, category := range body.Category {
		err = ValidateCategory(category, allowedCategories)
		if err != nil {
			validationErrors = append(validationErrors, web_helpers.NewFieldError("category", err.Error()))
		}
	}
	return validationErrors
}
