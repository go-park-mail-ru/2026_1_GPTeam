package models

import (
	"time"
)

type RefreshTokenModel struct {
	Uuid      string
	UserId    int
	ExpiredAt time.Time
	DeviceId  string
}
