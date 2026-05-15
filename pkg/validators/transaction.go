package validators

import "github.com/go-park-mail-ru/2026_1_GPTeam/internal/web/web_helpers"

func ValidateTransaction(body web_helpers.TransactionRequest, transactionTypes []string, categoryTypes []string) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := ValidateTransactionAccountId(body.AccountId)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("account_id", err.Error()))
	}
	err = validateTransactionTitle(body.Title)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", err.Error()))
	}
	err = validateTransactionDescription(body.Description)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", err.Error()))
	}
	err = validateTransactionValue(body.Value)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("value", err.Error()))
	}
	err = validateTransactionType(body.Type, transactionTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("type", err.Error()))
	}
	err = validateCategory(body.Category, categoryTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("category", err.Error()))
	}
	return validationErrors
}

func ValidateTransactionDraft(body web_helpers.TransactionRequest, transactionTypes []string, categoryTypes []string) []web_helpers.FieldError {
	var validationErrors []web_helpers.FieldError
	err := validateTransactionTitle(body.Title)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("title", err.Error()))
	}
	err = validateTransactionDescription(body.Description)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("description", err.Error()))
	}
	err = validateTransactionValue(body.Value)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("value", err.Error()))
	}
	err = validateTransactionType(body.Type, transactionTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("type", err.Error()))
	}
	err = validateCategory(body.Category, categoryTypes)
	if err != nil {
		validationErrors = append(validationErrors, web_helpers.NewFieldError("category", err.Error()))
	}
	return validationErrors
}

func ValidateImportFileColumns(firstLine []string) string {
	gpteamFile := []string{"Название", "Сумма", "Счёт", "Тип", "Категория", "Дата", "Описание"}
	sberFile := []string{"Дата", "Категория", "Сумма", "Остаток"}
	if err := checkSlicesEquals(firstLine, gpteamFile); err == nil {
		return "gpteam"
	}
	if err := checkSlicesEquals(firstLine, sberFile); err == nil {
		return "sber"
	}
	return ""
}
