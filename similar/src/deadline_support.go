package src

import "time"

type Duration = time.Duration
type Instant = time.Time

// Checks if a deadline was exeeded.
func deadline_exceeded(deadline Option[Instant]) bool {
	if deadline.Valid {
		deadline := deadline.Value
		return time.Now().After(deadline)
	} else {
		return false
	}
}

// Converst a duration into a deadline
func duration_to_deadline(add Duration) Option[Instant] {
	return Some(time.Now().Add(add))
}
