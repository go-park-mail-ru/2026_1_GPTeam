package validators

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	ServerError = fmt.Errorf("ошибка сервера")

	UsernameShortError        = fmt.Errorf("логин должен быть минимум 3 символа")
	UsernameWrongSymbolsError = fmt.Errorf("логин должен содержать только буквы латинского алфавита или цифры")

	PasswordShortError          = fmt.Errorf("пароль должен быть минимум 8 символов")
	PasswordsHasNoUpper         = fmt.Errorf("в пароле нет заглавной буквы")
	PasswordHasNoLower          = fmt.Errorf("в пароле нет строчной буквы")
	PasswordHasNoDigit          = fmt.Errorf("в пароле нет цифры")
	PasswordHasIncorrectSymbols = fmt.Errorf("пароль должен содержать только буквы латинского алфавита и цифры")

	EmailError = fmt.Errorf("некорректный адрес электронной почты")

	CurrencyNotAllowed = fmt.Errorf("валюта не поддерживается")

	TargetIsNegativeError = fmt.Errorf("планируемый бюджет не может быть меньше 0")
	TargetIsBigError      = fmt.Errorf("значение не может быть больше 1e18")

	StartDateInPastError = fmt.Errorf("дата начала не может быть в прошлом")
	EndDateInPastError   = fmt.Errorf("дата окончания должна быть позже даты начала")
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
		return PasswordShortError
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
		return PasswordsHasNoUpper
	}
	if !hasLower {
		return PasswordHasNoLower
	}
	if !hasDigit {
		return PasswordHasNoDigit
	}
	if hasInvalid {
		return PasswordHasIncorrectSymbols
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) >= 255 {
		return EmailError
	}
	matched, err := regexp.MatchString("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$", email)
	if err != nil {
		fmt.Println(err)
		return ServerError
	}
	if !matched {
		return EmailError
	}
	return nil
}

func ValidateCurrency(currency string) error {
	currency = strings.ToUpper(currency)
	allowedCurrencies := []string{
		"RUB",
		"USD",
		"EUR",
	}
	if !slices.Contains(allowedCurrencies, currency) {
		return CurrencyNotAllowed
	}
	return nil
}

func ValidateTargetBudget(target int) error {
	if target < 0 {
		return TargetIsNegativeError
	}
	if target > 1e18 {
		return TargetIsBigError
	}
	return nil
}

func ValidateStartDate(startDate time.Time) error {
	nowDate := time.Now()
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
