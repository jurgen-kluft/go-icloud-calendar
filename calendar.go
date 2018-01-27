package icalendar

import (
	"fmt"
	"sort"
	"time"
)

// Calendar is a structure that mainly contains events
type Calendar struct {
	Name               string
	Description        string
	URL                string
	Version            float64
	Timezone           time.Location
	Events             Events
	EventsByDate       map[string][]*Event
	EventsByID         map[string]*Event
	EventsByImportedID map[string]*Event
}

// Events is an array of Event
type Events []Event

func (events Events) Len() int {
	return len(events)
}

func (events Events) Less(i, j int) bool {
	return events[i].Start.Before(events[j].Start)
}

func (events Events) Swap(i, j int) {
	events[i], events[j] = events[j], events[i]
}

// New returns a new instance of a Calendar
func New() *Calendar {
	c := &Calendar{}
	c.Events = make([]Event, 0, 8)
	c.EventsByDate = make(map[string][]*Event)
	c.EventsByID = make(map[string]*Event)
	c.EventsByImportedID = make(map[string]*Event)
	return c
}

// SetEvent add event to the calendar
func (c *Calendar) SetEvent(event Event) (*Calendar, error) {

	// reference to the calendar
	if event.Owner == nil || event.Owner != c {
		event.Owner = c
	}
	// add the event to the main array with events
	c.Events = append(c.Events, event)

	// pointer to the added event in the main array
	eventPtr := &c.Events[len(c.Events)-1]

	// calculate the start and end day of the event
	eventStartTime := event.Start
	eventEndTime := event.End
	tz := c.Timezone
	eventStartDate := time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 0, 0, 0, 0, &tz)
	eventEndDate := time.Date(eventEndTime.Year(), eventEndTime.Month(), eventEndTime.Day(), 0, 0, 0, 0, &tz)

	// faster search by date, add each date from start to end date
	for eventDate := eventStartDate; eventDate.Before(eventEndDate) || eventDate.Equal(eventEndDate); eventDate = eventDate.Add(24 * time.Hour) {
		c.EventsByDate[eventDate.Format(YmdHis)] = append(c.EventsByDate[eventDate.Format(YmdHis)], eventPtr)
	}

	// faster search by id
	c.EventsByID[event.ID] = eventPtr

	if event.ImportedID != "" {
		c.EventsByImportedID[event.ImportedID] = eventPtr
	}

	return c, nil
}

// GetEventByID get event by id
func (c *Calendar) GetEventByID(eventID string) (*Event, error) {
	event, ok := c.EventsByID[eventID]
	if ok {
		return event, nil
	}
	return nil, fmt.Errorf("There is no event with id %s", eventID)
}

// GetEventByImportedID get event by imported id
func (c *Calendar) GetEventByImportedID(eventID string) (*Event, error) {
	event, ok := c.EventsByImportedID[eventID]
	if ok {
		return event, nil
	}
	return nil, fmt.Errorf("There is no event with id %s", eventID)
}

// GetEventsByDate get all events for specified date
func (c *Calendar) GetEventsByDate(dateTime time.Time) ([]*Event, error) {
	tz := c.Timezone
	day := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, &tz)
	events, ok := c.EventsByDate[day.Format(YmdHis)]
	if ok {
		return events, nil
	}
	return nil, fmt.Errorf("There are no events for the day %s", day.Format(YmdHis))
}

// GetUpcomingEvents returns the next n-Events.
func (c *Calendar) GetUpcomingEvents(n int) []Event {
	upcomingEvents := []Event{}

	// sort events of calendar
	sort.Sort(c.Events)

	now := time.Now()
	// find next event
	for _, event := range c.Events {
		if event.Start.After(now) {
			upcomingEvents = append(upcomingEvents, event)
			// break if we collect enough events
			if len(upcomingEvents) >= n {
				break
			}
		}
	}

	return upcomingEvents
}

func (c *Calendar) String() string {
	eventsCount := len(c.Events)
	name := c.Name
	desc := c.Description
	url := c.URL
	return fmt.Sprintf("Calendar %s about %s has %d events. Downloaded from %s .", name, desc, eventsCount, url)
}