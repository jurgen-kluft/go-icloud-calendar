package icalendar

import (
	"time"
)

// Timeline will contain events for a certain period
type Timeline struct {
	Start  time.Time
	End    time.Time
	Cal    *Calendar
	Events map[string][]Index
}
