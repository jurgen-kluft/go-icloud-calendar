package icalendar

import (
	"crypto/md5"
	"fmt"
	"time"
)

// Event holds all information for a Calendar Event
type Event struct {
	Start           time.Time
	End             time.Time
	Created         time.Time
	Modified        time.Time
	AlarmTime       time.Duration
	ImportedID      string
	Status          string
	Description     string
	Location        string
	Geo             *Geo
	Summary         string
	Rrule           string
	Class           string
	ID              string
	Sequence        int
	Attendees       []*Attendee
	Organizer       *Attendee
	IsWholeDayEvent bool
	Owner           *Calendar
	AlarmCallback   func(*Event)
}

// NewEvent will create a new instance of Event
func NewEvent() *Event {
	e := &Event{}
	e.Attendees = []*Attendee{}
	return e
}

// AddAttendee will add an Attendee to this Event
func (e *Event) AddAttendee(a *Attendee) {
	e.Attendees = append(e.Attendees, a)
}

// GenerateUUID generates an unique id for the event
func (e *Event) GenerateUUID() string {
	var toBeHashed string
	if e.ImportedID != "" {
		toBeHashed = fmt.Sprintf("%s%s%s", e.Start, e.End, e.ImportedID)
	} else {
		toBeHashed = fmt.Sprintf("%s%s%d", e.Start, e.End, time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", md5.Sum(stringToByte(toBeHashed)))
}

func (e *Event) String() string {
	from := e.Start.Local().Format(YmdHis)
	to := e.End.Local().Format(YmdHis)
	summ := e.Summary
	status := e.Status
	attendeeCount := len(e.Attendees)
	return fmt.Sprintf("Event(%s) from %s to %s about %s . %d people are invited to it", status, from, to, summ, attendeeCount)
}
