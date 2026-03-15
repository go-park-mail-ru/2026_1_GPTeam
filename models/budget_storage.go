package models

import (
	"time"
)

type BudgetInfo struct {
	Id          int
	Title       string
	Description string
	CreatedAt   time.Time
	StartAt     time.Time
	EndAt       time.Time
	Actual      int
	Target      int
	Currency    string
	Author      int
}
