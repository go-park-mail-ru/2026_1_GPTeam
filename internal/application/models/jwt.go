package models

//go:generate easyjson -all models.go
import (
	"time"
)

type RefreshTokenModel struct {
	Uuid      string
	UserId    int
	ExpiredAt time.Time
	DeviceId  string
}
