package currency_converter

func ConvertToRub(value float64, currency string) float64 {
	switch currency {
	case "USD":
		return value * 75
	case "EUR":
		return value * 88
	default:
		return value
	}
}
