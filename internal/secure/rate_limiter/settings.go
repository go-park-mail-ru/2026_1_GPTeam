package rate_limiter

import "time"

const RefillRate = 1
const MaxCount = 100
const TTL = 24 * time.Hour
const BlockDuration = 15 * time.Minute
const CleanInterval = 24 * time.Hour
