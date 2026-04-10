package rate_limiter

import "time"

const RefillRate = 1
const MaxCount = 100
const TTL = 86400 // 1 day
const BlockDuration = 15 * time.Minute
