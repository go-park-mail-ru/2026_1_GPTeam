package validators

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
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
	currency = strings.ToUpper(currency)
	if !slices.Contains(allowedCurrencies, currency) {
		return CurrencyNotAllowed
	}
	return nil
}

func ValidateTargetBudget(target int) error {
	if target <= 0 {
		return TargetIsNegativeError
	}
	if target > 1e18 {
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
