package models

import "time"

type SupportModel struct {
	Id        int
	UserId    int
	Category  string
	Message   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
}
