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

func ValidateCurrency(currency string, allowedCurrencies []string) error {
	if !slices.Contains(allowedCurrencies, currency) {
		return CurrencyNotAllowedError
	}
	return nil
}

func ValidateTargetBudget(target int) error {
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

func ValidateStartDate(startDate time.Time) error {
	nowTime := time.Now()
	nowDate := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, startDate.Location())
	if startDate.Before(nowDate) {
		return StartDateInPastError
	}
	return nil
}

func ValidateEndDate(startDate time.Time, endDate time.Time) error {
	if endDate.IsZero() {
		return nil
	}
	if endDate.Before(startDate) {
		return EndDateInPastError
	}
	return nil
}

func validateTransactionTitle(title string) error {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) == 0 {
		return errors.New("название не может быть пустым")
	}
	if utf8.RuneCountInString(title) > 255 {
		return errors.New("название не должно превышать 255 символов")
	}
	return nil
}

func validateTransactionDescription(description string) error {
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) == 0 {
		return errors.New("описание не может быть пустым")
	}
	return nil
}

func validateTransactionValue(value float64) error {
	if value <= 0 {
		return errors.New("сумма должна быть больше 0")
	}
	if value > 1_000_000_000 {
		return errors.New("сумма не может превышать 1 000 000 000")
	}
	return nil
}

func validateTransactionType(transactionType string, allowedTypes []string) error {
	for _, t := range allowedTypes {
		if t == transactionType {
			return nil
		}
	}
	return errors.New("недопустимый тип транзакции")
}

func validateTransactionCategory(category string, allowedCategories []string) error {
	for _, c := range allowedCategories {
		if c == category {
			return nil
		}
	}
	return errors.New("недопустимая категория")
}
