package models

import "time"

type AccountModel struct {
	Id        int
	Name      string
	Balance   float64
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
