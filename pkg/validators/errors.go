package validators

import "fmt"

var (
	ErrMinLength = fmt.Errorf("текст слишком короткий")
	ErrMaxLength = fmt.Errorf("текст слишком длинный")
	ErrNoEquals  = fmt.Errorf("значения не совпадают")

	ErrServerError = fmt.Errorf("ошибка сервера")

	ErrUsernameShort        = fmt.Errorf("логин должен быть минимум 3 символа")
	ErrUsernameWrongSymbols = fmt.Errorf("логин должен содержать только буквы латинского алфавита или цифры")
	ErrIncorrectPassword    = fmt.Errorf("пароль должен содержать заглавные, строчные буквы латинского алфавита и цифры (не менее 8 символов)")
	ErrPasswordsNotSame     = fmt.Errorf("пароли не совпадают")
	ErrEmailError           = fmt.Errorf("некорректный адрес электронной почты")

	ErrBudgetTitle            = fmt.Errorf("заголовок пустой")
	ErrBudgetTitleTooLong     = fmt.Errorf("заголовок не может быть длиннее 255 символов")
	ErrBudgetDescriptionEmpty = fmt.Errorf("описание пустое")
	ErrCurrencyNotAllowed     = fmt.Errorf("валюта не поддерживается")
	ErrTargetIsNegative       = fmt.Errorf("планируемый бюджет не может быть меньше 0")
	ErrTargetIsZero           = fmt.Errorf("планируемый бюджет не может быть равен 0")
	ErrTargetIsBig            = fmt.Errorf("значение не может быть больше 1e18")
	ErrStartDateInPast        = fmt.Errorf("дата начала не может быть в прошлом")
	ErrEndDateInPast          = fmt.Errorf("дата окончания должна быть позже даты начала")

	ErrValueIsNegative = fmt.Errorf("значение не может быть меньше нуля")
	ErrValueIsBig      = fmt.Errorf("значение не может быть больше 1e12")

	ErrTransactionTypeNotAllowed   = fmt.Errorf("тип транзакции не поддерживается")
	ErrCategoryNotAllowed          = fmt.Errorf("категория не поддерживается")
	ErrTransactionTitleEmpty       = fmt.Errorf("заголовок пустой")
	ErrTransactionTitleLong        = fmt.Errorf("заголовок не может быть длиннее 255 символов")
	ErrTransactionDescriptionEmpty = fmt.Errorf("описание пустое")
)
