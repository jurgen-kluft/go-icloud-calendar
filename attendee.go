package icalendar

import (
	"fmt"
)

// Attendee takes part in a calendar event
type Attendee struct {
	Name   string
	Email  string
	Status string
	Role   string
	Type   string
}

// NewAttendee will create a new Attendee instance
func NewAttendee() *Attendee {
	a := new(Attendee)
	return a
}

func (a *Attendee) String() string {
	return fmt.Sprintf("%s with email %s", a.Name, a.Email)
}
