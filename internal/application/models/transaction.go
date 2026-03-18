package models

import "time"

type TransactionModel struct {
	Id              int
	UserId          int
	AccountId       int
	Value           int
	Type            string
	Category        string
	Title           string
	Description     string
	CreatedAt       time.Time
	TransactionDate time.Time
}
