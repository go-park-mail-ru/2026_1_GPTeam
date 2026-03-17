package validators

import "fmt"

var (
	ServerError = fmt.Errorf("ошибка сервера")

	UsernameShortError        = fmt.Errorf("логин должен быть минимум 3 символа")
	UsernameWrongSymbolsError = fmt.Errorf("логин должен содержать только буквы латинского алфавита или цифры")

	IncorrectPasswordError = fmt.Errorf("пароль должен содержать заглавные, строчные буквы латинского алфавита и цифры (не менее 8 символов)")

	EmailError = fmt.Errorf("некорректный адрес электронной почты")

	CurrencyNotAllowed = fmt.Errorf("валюта не поддерживается")

	TargetIsNegativeError = fmt.Errorf("планируемый бюджет не может быть меньше 0")
	TargetIsBigError      = fmt.Errorf("значение не может быть больше 1e18")

	StartDateInPastError = fmt.Errorf("дата начала не может быть в прошлом")
	EndDateInPastError   = fmt.Errorf("дата окончания должна быть позже даты начала")
)
