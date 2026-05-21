package models

//go:generate easyjson -all models.go
import "time"

type BucketModel struct {
	Count          int
	LastRefillTime time.Time
	BlockedUntil   time.Time
	LastSeen       time.Time
}

type PermanentBlockedIps struct {
	Ips []string
}
