package models

import (
	"time"
)

type RefreshTokenInfo struct {
	UserID    string
	ExpiredAt time.Time
	DeviceID  string
}
