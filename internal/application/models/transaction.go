package models

import "time"

type TransactionModel struct {
	Id              int
	UserId          int
	AccountId       int
	Value           float64
	Type            string
	Category        string
	Title           string
	Description     string
	CreatedAt       time.Time
	TransactionDate time.Time
	DeletedAt       time.Time
	UpdatedAt       time.Time
}

type TransactionDraft struct {
	RawText     string    `json:"raw_text"`
	Value       float64   `json:"value"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	RecordedAt  time.Time `json:"recorded_at"`
	Date        time.Time `json:"date"`
}
