package validators

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// User section

func validateUsername(username string) error {
	if len(username) < 3 {
		return ErrUsernameShort
	}
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", username)
	if err != nil {
		fmt.Println(err)
		return ErrServerError
	}
	if !matched {
		return ErrUsernameWrongSymbols
	}
	return nil
}

func validatePassword(passwordStr string) error {
	password := []rune(passwordStr)
	if len(password) < 8 {
		return ErrIncorrectPassword
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
		return ErrIncorrectPassword
	}
	if !hasLower {
		return ErrIncorrectPassword
	}
	if !hasDigit {
		return ErrIncorrectPassword
	}
	if hasInvalid {
		return ErrIncorrectPassword
	}
	return nil
}

func validateConfirmPassword(passwordStr string, confirmPasswordStr string) error {
	if passwordStr != confirmPasswordStr {
		return ErrPasswordsNotSame
	}
	return nil
}

func validateEmail(email string) error {
	if len(email) == 0 || len(email) >= 255 {
		return ErrEmailError
	}
	matched, err := regexp.MatchString("^[A-Za-zа-яёА-ЯЁ0-9._%+-]+@[A-Za-zа-яёА-ЯЁ0-9.-]+\\.[A-Za-zа-яёА-ЯЁ]{2,}$", email)
	if err != nil {
		fmt.Println(err)
		return ErrServerError
	}
	if !matched {
		return ErrEmailError
	}
	return nil
}

// Budget section

func validateBudgetTitle(title string) error {
	if len(title) == 0 {
		return ErrBudgetTitle
	}
	if len(title) > 255 {
		return ErrBudgetTitleTooLong
	}
	return nil
}

func validateBudgetDescription(description string) error {
	if len(description) == 0 {
		return ErrBudgetDescriptionEmpty
	}
	return nil
}

func validateCurrency(currency string, allowedCurrencies []string) error {
	if !slices.Contains(allowedCurrencies, currency) {
		return ErrCurrencyNotAllowed
	}
	return nil
}

func validateTargetBudget(target float64) error {
	if target < 0 {
		return ErrTargetIsNegative
	}
	if target == 0 {
		return ErrTargetIsZero
	}
	if target > 1_000_000_000 {
		return ErrTargetIsBig
	}
	return nil
}

func validateActualBudget(actual int) error {
	if actual < 0 {
		return ErrValueIsNegative
	}
	if actual > 1_000_000_000 {
		return ErrValueIsBig
	}
	return nil
}

func validateBudgetStartDate(startDate time.Time) error {
	nowTime := time.Now()
	nowDate := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, startDate.Location())
	if startDate.Before(nowDate) {
		return ErrStartDateInPast
	}
	return nil
}

func validateBudgetEndDate(startDate time.Time, endDate time.Time) error {
	if endDate.IsZero() {
		return nil
	}
	if endDate.Before(startDate) {
		return ErrEndDateInPast
	}
	return nil
}

// Transaction section

func validateTransactionTitle(title string) error {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) == 0 {
		return ErrTransactionTitleEmpty
	}
	if utf8.RuneCountInString(title) > 255 {
		return ErrTransactionTitleLong
	}
	return nil
}

func validateTransactionDescription(description string) error {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) == 0 {
		return ErrTransactionDescriptionEmpty
	}
	return nil
}

func validateTransactionValue(value float64) error {
	if value <= 0 {
		return ErrValueIsNegative
	}
	if value > 1_000_000_000 {
		return ErrValueIsBig
	}
	return nil
}

func validateTransactionType(transactionType string, allowedTypes []string) error {
	for _, t := range allowedTypes {
		if t == transactionType {
			return nil
		}
	}
	return ErrTransactionTypeNotAllowed
}

func validateCategory(category string, allowedCategories []string) error {
	for _, c := range allowedCategories {
		if c == category {
			return nil
		}
	}
	return ErrCategoryNotAllowed
}

func validateLength(text string, minLength int, maxLength int) error {
	if len(text) < minLength {
		return ErrMinLength
	}
	if len(text) > maxLength {
		return ErrMaxLength
	}
	return nil
}

func checkEqual[T comparable](a, b T) error {
	if a != b {
		return ErrNoEquals
	}
	return nil
}

func ValidateTransactionAccountId(accountId int) error {
	if accountId <= 0 {
		return errors.New("account_id must be positive")
	}
	return nil
}

func checkSlicesEquals[T comparable](a, b []T) error {
	if len(a) != len(b) {
		return ErrNoEquals
	}
	for i := range a {
		if a[i] != b[i] {
			return ErrNoEquals
		}
	}
	return nil
}
