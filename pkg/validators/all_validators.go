package validators

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// User section

func ValidateUsername(username string) error {
	if len(username) < 3 {
		return UsernameShortError
	}
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", username)
	if err != nil {
		fmt.Println(err)
		return ServerError
	}
	if !matched {
		return UsernameWrongSymbolsError
	}
	return nil
}

func ValidatePassword(passwordStr string) error {
	password := []rune(passwordStr)
	if len(password) < 8 {
		return IncorrectPasswordError
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasInvalid := false
	for i := 0; i < len(password); i++ {
		if 'A' <= password[i] && password[i] <= 'Z' {
			hasUpper = true
		} else if 'a' <= password[i] && password[i] <= 'z' {
			hasLower = true
		} else if '0' <= password[i] && password[i] <= '9' {
			hasDigit = true
		} else {
			hasInvalid = true
		}
	}
	if !hasUpper {
		return IncorrectPasswordError
	}
	if !hasLower {
		return IncorrectPasswordError
	}
	if !hasDigit {
		return IncorrectPasswordError
	}
	if hasInvalid {
		return IncorrectPasswordError
	}
	return nil
}

func ValidateConfirmPassword(passwordStr string, confirmPasswordStr string) error {
	if passwordStr != confirmPasswordStr {
		return PasswordsNotSameError
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) >= 255 {
		return EmailError
	}
	matched, err := regexp.MatchString("^[A-Za-zа-яёА-ЯЁ0-9._%+-]+@[A-Za-zа-яёА-ЯЁ0-9.-]+\\.[A-Za-zа-яёА-ЯЁ]{2,}$", email)
	if err != nil {
		fmt.Println(err)
		return ServerError
	}
	if !matched {
		return EmailError
	}
	return nil
}

// Budget section

func validateBudgetTitle(title string) error {
	if len(title) == 0 {
		return BudgetTitleEmpty
	}
	if len(title) > 255 {
		return BudgetTitleTooLong
	}
	return nil
}

func validateBudgetDescription(description string) error {
	if len(description) == 0 {
		return BudgetDescriptionEmpty
	}
	return nil
}

func validateCurrency(currency string, allowedCurrencies []string) error {
	if !slices.Contains(allowedCurrencies, currency) {
		return CurrencyNotAllowedError
	}
	return nil
}

func validateTargetBudget(target int) error {
	if target < 0 {
		return TargetIsNegativeError
	}
	if target == 0 {
		return TargetIsZeroError
	}
	if target > 1e12 {
		return TargetIsBigError
	}
	return nil
}

func validateActualBudget(actual int) error {
	if actual < 0 {
		return ValueIsNegativeError
	}
	if actual > 1e12 {
		return ValueIsBigError
	}
	return nil
}

func validateBudgetStartDate(startDate time.Time) error {
	nowTime := time.Now()
	nowDate := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, startDate.Location())
	if startDate.Before(nowDate) {
		return StartDateInPastError
	}
	return nil
}

func validateBudgetEndDate(startDate time.Time, endDate time.Time) error {
	if endDate.IsZero() {
		return nil
	}
	if endDate.Before(startDate) {
		return EndDateInPastError
	}
	return nil
}

// Transaction section

func ValidateTransactionTitle(title string) error {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) == 0 {
		return TransactionTitleEmptyError
	}
	if utf8.RuneCountInString(title) > 255 {
		return TransactionTitleLongError
	}
	return nil
}

func ValidateTransactionDescription(description string) error {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) == 0 {
		return TransactionDescriptionEmptyError
	}
	return nil
}

func ValidateTransactionValue(value float64) error {
	if value <= 0 {
		return ValueIsNegativeError
	}
	if value > 1_000_000_000 {
		return ValueIsBigError
	}
	return nil
}

func ValidateTransactionType(transactionType string, allowedTypes []string) error {
	for _, t := range allowedTypes {
		if t == transactionType {
			return nil
		}
	}
	return TransactionTypeNotAllowedError
}

func validateCategory(category string, allowedCategories []string) error {
	for _, c := range allowedCategories {
		if c == category {
			return nil
		}
	}
	return CategoryNotAllowedError
}

func validateLength(text string, minLength int, maxLength int) error {
	if len(text) < minLength {
		return MinLengthError
	}
	if len(text) > maxLength {
		return MaxLengthError
	}
	return nil
}

func checkEqual[T comparable](a, b T) error {
	if a != b {
		return NoEqualsError
	}
	return nil
}
