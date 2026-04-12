package rate_limiter

import "time"

const RefillRateInHalfSecond = 1
const MaxCount = 100
const BlockDuration = 15 * time.Minute
