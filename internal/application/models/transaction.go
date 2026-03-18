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
}
