package models

import "time"

type ContextKey string

const UserContextKey ContextKey = "user"

type UserModel struct {
	Id        int
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
	LastLogin time.Time
	AvatarUrl string
	UpdatedAt time.Time
	Active    bool
}

type UpdateUserProfile struct {
	Id        int
	Username  *string
	Email     *string
	Password  *string
	AvatarUrl *string
	UpdatedAt time.Time
}
