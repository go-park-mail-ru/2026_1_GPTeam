package models

import (
	"time"
)

type RefreshTokenInfo struct {
	Uuid      string
	UserID    int
	ExpiredAt time.Time
	DeviceID  string
}
