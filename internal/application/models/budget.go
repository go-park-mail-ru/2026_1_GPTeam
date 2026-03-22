package models

import (
	"time"
)

type BudgetModel struct {
	Id          int
	Title       string
	Description string
	CreatedAt   time.Time
	StartAt     time.Time
	EndAt       time.Time
	UpdatedAt   time.Time
	Actual      float64
	Target      float64
	Currency    string
	Author      int
	Active      bool
}
