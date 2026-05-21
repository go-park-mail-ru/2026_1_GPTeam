package models

//go:generate easyjson -all models.go
import "time"

type AccountModel struct {
	Id        int
	Name      string
	Balance   float64
	Currency  string
	OwnerId   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccountCreateModel struct {
	Name     string
	Balance  float64
	Currency string
	OwnerId  int
}

type AccountUpdateModel struct {
	Name     *string
	Balance  *float64
	Currency *string
}
