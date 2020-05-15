package icalendar

import (
	"fmt"
	"github.com/jurgen-kluft/go-icloud-calendar/rrule"
	"math"
	"time"
)

// Calendar is a structure that mainly contains events
type Calendar struct {
	Name                string
	Description         string
	reader              Reader
	parser              *parser
	Version             float64
	Timezone            time.Location
	Events              Events
	EventsByDate        map[string][]Index
	EventsByID          map[string]Index
	EventsByImportedID  map[string]Index
	RecurringEvents     []Index
	RecurringEventRules RRules
}

type Index int

func (i Index) IsValid() bool {
	return int(i) != math.MaxInt32
}
func (i Index) Invalid() Index {
	return Index(math.MaxInt32)
}
func (i Index) ToInt() int {
	return int(i)
}

// Events is an array of Event
type Events []*Event
type RRules []*rrule.RRule

func (events Events) Len() int {
	return len(events)
}

func (events Events) Less(i, j int) bool {
	return events[i].Start.Before(events[j].Start)
}

func (events Events) Swap(i, j int) {
	events[i], events[j] = events[j], events[i]
}

func newCalendar() *Calendar {
	c := &Calendar{}
	c.reader = nil
	c.parser = nil
	c.Events = make([]*Event, 0, 8)
	c.EventsByDate = make(map[string][]Index)
	c.EventsByID = make(map[string]Index)
	c.EventsByImportedID = make(map[string]Index)
	c.RecurringEvents = make([]Index, 0, 0)
	c.RecurringEventRules = make([]*rrule.RRule, 0, 0)
	return c
}

// NewURLCalendar returns a new instance of a Calendar that has a URL source
func NewURLCalendar(URL string) *Calendar {
	c := newCalendar()
	c.reader = readingFromURL(URL)
	c.parser = createParser(c.reader)
	return c
}

// NewFileCalendar returns a new instance of a Calendar that has a file source
func NewFileCalendar(filepath string) *Calendar {
	c := newCalendar()
	c.reader = readingFromFile(filepath)
	c.parser = createParser(c.reader)
	return c
}

func (c *Calendar) Load() error {
	calendar := newCalendar()
	calendar.parser = c.parser
	calendar.reader = c.reader
	err := c.parser.read(calendar)
	if err == nil {
		// Take content of loaded calendar
		c.Name = calendar.Name
		c.Description = calendar.Description
		c.reader = calendar.reader
		c.parser = calendar.parser
		c.Version = calendar.Version
		c.Timezone = calendar.Timezone
		c.Events = calendar.Events
		c.EventsByDate = calendar.EventsByDate
		c.EventsByID = calendar.EventsByID
		c.EventsByImportedID = calendar.EventsByImportedID
	}

	// Sort ?

	return err
}

// InsertEvent add event to the calendar
func (c *Calendar) InsertEvent(event *Event) (err error) {

	// reference to the calendar
	if event.Owner == nil || event.Owner != c {
		event.Owner = c
	}

	// add the event to the main array with events
	eventRef := len(c.Events)
	c.Events = append(c.Events, event)

	if event.Rrule == "" {

		// calculate the start and end day of the event
		eventStartTime := event.Start
		eventEndTime := event.End
		tz := c.Timezone
		eventStartDate := time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 0, 0, 0, 0, &tz)
		eventEndDate := time.Date(eventEndTime.Year(), eventEndTime.Month(), eventEndTime.Day(), 0, 0, 0, 0, &tz)

		// faster search by date, add each date from start to end date
		for eventDate := eventStartDate; eventDate.Before(eventEndDate) || eventDate.Equal(eventEndDate); eventDate = eventDate.Add(24 * time.Hour) {
			c.EventsByDate[eventDate.Format(YmdHis)] = append(c.EventsByDate[eventDate.Format(YmdHis)], Index(eventRef))
		}

		// faster search by id
		c.EventsByID[event.ID] = Index(eventRef)

		if event.ImportedID != "" {
			c.EventsByImportedID[event.ImportedID] = Index(eventRef)
		}

	} else {
		var rule *rrule.RRule
		rule, err = rrule.StrToRRule(event.Rrule)
		if err == nil {
			c.RecurringEvents = append(c.RecurringEvents, Index(eventRef))
			c.RecurringEventRules = append(c.RecurringEventRules, rule)
			err = rule.Compile(event.Start, event.End)
		}

		// faster search by id
		c.EventsByID[event.ID] = Index(-eventRef)

		if event.ImportedID != "" {
			c.EventsByImportedID[event.ImportedID] = Index(-eventRef)
		}

	}

	return err
}

// GetEventByIndex get event by index
func (c *Calendar) GetEventByIndex(e Index) (*Event, error) {
	i := e.ToInt()
	if i < 0 {
		i = -i
		if i < len(c.RecurringEvents) {
			i = c.RecurringEvents[i].ToInt()
		} else {
			return nil, fmt.Errorf("There is no recurring event for index %d", i)
		}
	}
	if i < len(c.Events) {
		return c.Events[i], nil
	}
	return nil, fmt.Errorf("There is no event for index %d", i)
}

// GetEventIndexByID get event by id
func (c *Calendar) GetEventIndexByID(eventID string) (Index, error) {
	event, ok := c.EventsByID[eventID]
	if ok {
		return event, nil
	}
	return event.Invalid(), fmt.Errorf("There is no event with id %s", eventID)
}

// GetEventIndexByImportedID get event by imported id
func (c *Calendar) GetEventIndexByImportedID(eventID string) (Index, error) {
	event, ok := c.EventsByImportedID[eventID]
	if ok {
		return event, nil
	}
	return event.Invalid(), fmt.Errorf("There is no event with id %s", eventID)
}

// GetEventIndicesByDate get all events for specified date
func (c *Calendar) GetEventIndicesByDate(dateTime time.Time) []Index {
	tz := c.Timezone
	day := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, &tz)
	events, ok := c.EventsByDate[day.Format(YmdHis)]
	if ok {
		return events
	}
	return []Index{}
}

// GetEventsFor get all active events for specified date
func (c *Calendar) GetEventsFor(dateTime time.Time) []*Event {
	tz := c.Timezone
	day := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, &tz)
	today := []*Event{}
	events, ok := c.EventsByDate[day.Format(YmdHis)]
	if ok {
		for _, i := range events {
			event, err := c.GetEventByIndex(i)
			if err == nil {
				today = append(today, event)
			}
		}
	}
	for i, rer := range c.RecurringEventRules {
		if rer.Includes(dateTime) {
			rei := c.RecurringEvents[i]
			event, err := c.GetEventByIndex(rei)
			if err == nil {
				today = append(today, event)
			}
		}
	}

	return today
}

func (c *Calendar) String() string {
	eventsCount := len(c.Events)
	name := c.Name
	desc := c.Description
	return fmt.Sprintf("Calendar %s about %s has %d events.", name, desc, eventsCount)
}
