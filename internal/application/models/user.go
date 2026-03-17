package models

import (
	"time"
)

type UserModel struct {
	Id        int
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	LastLogin time.Time
	AvatarUrl string
	UpdatedAt time.Time
}
