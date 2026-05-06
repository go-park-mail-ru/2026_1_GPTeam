package groq

import "time"

type TransactionDraft struct {
	RawText     string
	Value       float64
	Type        string
	Category    string
	Title       string
	Description string
	Date        time.Time
}
