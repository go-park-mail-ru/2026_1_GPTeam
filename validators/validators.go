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
		return fmt.Errorf("логин должен быть минимум 3 символа")
	}
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", username)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("ошибка сервера")
	}
	if !matched {
		return fmt.Errorf("логин должен содержать только буквы латинского алфавита или цифры")
	}
	return nil
}

func ValidatePassword(passwordStr string) error {
	password := []rune(passwordStr)
	if len(password) < 8 {
		return fmt.Errorf("пароль должен быть минимум 8 символов")
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
		return fmt.Errorf("в пароле нет заглавной буквы")
	}
	if !hasLower {
		return fmt.Errorf("в пароле нет строчной буквы")
	}
	if !hasDigit {
		return fmt.Errorf("в пароле нет цифры")
	}
	if hasInvalid {
		return fmt.Errorf("пароль должен содержать только буквы латинского алфавита и цифры")
	}
	return nil
}

func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) >= 255 {
		return fmt.Errorf("некорректный адрес электронной почты")
	}
	matched, err := regexp.MatchString("^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,}$", email)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("ошибка сервера")
	}
	if !matched {
		return fmt.Errorf("некорректный адрес электронной почты")
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
		return fmt.Errorf("валюта не поддерживается")
	}
	return nil
}

func ValidateTargetBudget(target int) error {
	if target < 0 {
		return fmt.Errorf("планируемый бюджет не может быть меньше 0")
	}
	if target > 1e18 {
		return fmt.Errorf("значение не может быть больше 1e18")
	}
	return nil
}

func ValidateStartDate(startDate time.Time) error {
	nowDate := time.Now()
	if startDate.Before(nowDate) {
		return fmt.Errorf("дата начала не может быть в прошлом")
	}
	return nil
}

func ValidateEndDate(startDate time.Time, endDate time.Time) error {
	if endDate.IsZero() {
		return nil
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("дата окончания должна быть позже даты начала")
	}
	return nil
}
