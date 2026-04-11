package models

import "time"

type BucketModel struct {
	Count          int
	LastRefillTime time.Time
	BlockedUntil   time.Time
	LastSeen       time.Time
}
