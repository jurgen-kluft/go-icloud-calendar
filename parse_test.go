package icalendar

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestLoadCalendar(t *testing.T) {
	reader := readingFromFile("testCalendars/2eventsCal.ics")
	parser := createParser(reader)

	calendar := newCalendar()
	err := parser.read(calendar)

	if err != nil {
		parseErrors := parser.getErrors()
		for i, pErr := range parseErrors {
			t.Errorf("Parsing Error №%d: %s", i, pErr)
		}
	}
}

func TestLoadCalendar2(t *testing.T) {
	reader := readingFromFile("testCalendars/4eventsWithRRule.ics")
	parser := createParser(reader)

	calendar := newCalendar()
	err := parser.read(calendar)

	if err != nil {
		parseErrors := parser.getErrors()
		for i, pErr := range parseErrors {
			t.Errorf("Parsing Error %d: %s", i, pErr)
		}
	}

	tl := calendar.GetTimeline(time.Now(), 1)
	t.Logf("Timeline with %d days", len(tl.Events))
	for day, events := range tl.Events {
		t.Logf("Day %s; recurring events: %d", day, len(events))
		for _, e := range events {
			event, err := calendar.GetEventByIndex(e)
			if err != nil {
				t.Errorf("   Event by index error %s", err)
			} else {
				t.Logf("   Recurring event: %s", event.String())
			}
		}
	}
}
func TestNewParser(t *testing.T) {
	reader := readingFromFile("testCalendars/2eventsCal.ics")
	parser := createParser(reader)
	rType := fmt.Sprintf("%v", reflect.TypeOf(parser))
	if rType != "*icalendar.parser" {
		t.Errorf("Failed to create *icalendar.Parser !")
	}
}

func TestParsingNotExistingCalendar(t *testing.T) {
	reader := readingFromFile("testCalendars/notFound.ics")
	parser := createParser(reader)
	calendar := newCalendar()
	parser.read(calendar)
	parseErrors := parser.getErrors()
	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}
}

func TestParsingWrongCalendarUrls(t *testing.T) {
	reader := readingFromURL("http://localhost/goTestFails")
	parser := createParser(reader)
	calendar := newCalendar()
	err := parser.read(calendar)
	parseErrors := parser.getErrors()

	if err == nil {
		t.Errorf("Expected 1 error, got none.\n")
	}

	if len(parseErrors) != 1 {
		t.Errorf("Expected 1 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	if len(calendar.Events) != 0 {
		t.Errorf("Expected a calendar with 0 events, found %d events", len(calendar.Events))
	}
}

func TestCalendarInfo(t *testing.T) {
	reader := readingFromFile("testCalendars/2eventsCal.ics")
	parser := createParser(reader)
	calendar := newCalendar()
	parser.read(calendar)
	parseErrors := parser.getErrors()

	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n %#v", len(parseErrors), parseErrors)
	}

	if calendar.Name != "2 Events Cal" {
		t.Errorf("Expected name '%s' calendar, got '%s' calendars", "2 Events Cal", calendar.Name)
	}

	if calendar.Description != "The cal has 2 events(1st with attendees and second without)" {
		t.Errorf("Expected description '%s' calendar, got '%s' calendars", "The cal has 2 events(1st with attendees and second without)", calendar.Description)
	}

	if calendar.Version != 2.0 {
		t.Errorf("Expected version %v calendar, got %v calendars", 2.0, calendar.Version)
	}

	events := calendar.Events
	if len(events) != 2 {
		t.Errorf("Expected %d events in calendar, got %d events", 2, len(events))
	}

	eventsByDates := calendar.EventsByDate
	if len(eventsByDates) != 2 {
		t.Errorf("Expected %d events by date in calendar, got %d events", 2, len(eventsByDates))
	}

	geometryExamIcsFormat, errICS := time.Parse(IcsFormat, "20140616T060000Z")
	if errICS != nil {
		t.Errorf("(ics time format) Unexpected error %s", errICS)
	}

	geometryExamYmdHis, errYMD := time.Parse(YmdHis, "2014-06-16 06:00:00")
	if errYMD != nil {
		t.Errorf("(YmdHis time format) Unexpected error %s", errYMD)
	}
	eventsByDate := calendar.GetEventIndicesByDate(geometryExamIcsFormat)
	if len(eventsByDate) != 1 {
		t.Errorf("(ics time format) Expected %d events in calendar for the date 2014-06-16, got %d events", 1, len(eventsByDate))
	}

	eventsByDate = calendar.GetEventIndicesByDate(geometryExamYmdHis)
	if len(eventsByDate) != 1 {
		t.Errorf("(YmdHis time format) Expected %d events in calendar for the date 2014-06-16, got %d events", 1, len(eventsByDate))
	}

}

func TestCalendarEvents(t *testing.T) {
	reader := readingFromFile("testCalendars/2eventsCal.ics")
	parser := createParser(reader)
	calendar := newCalendar()
	parser.read(calendar)
	parseErrors := parser.getErrors()
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	ievent, err := calendar.GetEventIndexByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s", err)
	}

	event, err := calendar.GetEventByIndex(ievent)

	//  event must have
	start, _ := time.Parse(IcsFormat, "20140714T100000Z")
	end, _ := time.Parse(IcsFormat, "20140714T110000Z")
	created, _ := time.Parse(IcsFormat, "20140515T075711Z")
	modified, _ := time.Parse(IcsFormat, "20141125T074253Z")
	location := "In The Office"
	geo := NewGeo("39.620511", "-75.852557")
	desc := "1. Report on previous weekly tasks. \\n2. Plan of the present weekly tasks."
	seq := 1
	status := "CONFIRMED"
	summary := "General Operative Meeting"
	rrule := ""
	attendeesCount := 3

	org := new(Attendee)
	org.Name = ("r.chupetlovska@gmail.com")
	org.Email = ("r.chupetlovska@gmail.com")

	if event.Start != start {
		t.Errorf("Expected start %s, found %s", start, event.Start)
	}

	if event.End != end {
		t.Errorf("Expected end %s, found %s", end, event.End)
	}

	if event.Created != created {
		t.Errorf("Expected created %s, found %s", created, event.Created)
	}

	if event.Modified != modified {
		t.Errorf("Expected modified %s, found %s", modified, event.Modified)
	}

	if event.Location != location {
		t.Errorf("Expected location %s, found %s", location, event.Location)
	}

	if event.Geo.latStr != geo.latStr {
		t.Errorf("Expected geo %s, found %s", geo.latStr, event.Geo.latStr)
	}

	if event.Geo.longStr != geo.longStr {
		t.Errorf("Expected geo %s, found %s", geo.longStr, event.Geo.longStr)
	}

	if event.Description != desc {
		t.Errorf("Expected description %s, found %s", desc, event.Description)
	}

	if event.Sequence != seq {
		t.Errorf("Expected sequence %d, found %d", seq, event.Sequence)
	}

	if event.Status != status {
		t.Errorf("Expected status %s, found %s", status, event.Status)
	}

	if event.Summary != summary {
		t.Errorf("Expected status %s, found %s", summary, event.Summary)
	}

	if event.Rrule != rrule {
		t.Errorf("Expected rrule %s, found %s", rrule, event.Rrule)
	}

	if len(event.Attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %d, found %d", attendeesCount, len(event.Attendees))
	}

	eventOrg := event.Organizer
	if *eventOrg != *org {
		t.Errorf("Expected organizer %s, found %s", org, event.Organizer)
	}

	// SECOND EVENT WITHOUT ATTENDEES AND ORGANIZER
	ieventNoAttendees, errNoAttendees := calendar.GetEventIndexByImportedID("mhhesb7si5968njvthgbiub7nk@google.com")
	attendeesCount = 0
	org = new(Attendee)

	if errNoAttendees != nil {
		t.Errorf("Failed to get event by id with error %s", errNoAttendees)
	}
	eventNoAttendees, errNoAttendees := calendar.GetEventByIndex(ieventNoAttendees)
	if errNoAttendees != nil {
		t.Errorf("Failed to get event by id with error %s", errNoAttendees)
	}

	if len(eventNoAttendees.Attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %d, found %d", attendeesCount, len(event.Attendees))
	}

	if eventNoAttendees.Organizer != nil {
		t.Errorf("Expected organizer %s, found %s", org, eventNoAttendees.Organizer)
	}
}

func TestCalendarEventAttendees(t *testing.T) {
	reader := readingFromFile("testCalendars/2eventsCal.ics")
	parser := createParser(reader)
	calendar := newCalendar()
	parser.read(calendar)
	parseErrors := parser.getErrors()

	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	ievent, err := calendar.GetEventIndexByImportedID("btb9tnpcnd4ng9rn31rdo0irn8@google.com")
	if err != nil {
		t.Errorf("Failed to get event by id with error %s", err)
	}
	event, err := calendar.GetEventByIndex(ievent)

	attendees := event.Attendees
	attendeesCount := 3

	if len(attendees) != attendeesCount {
		t.Errorf("Expected attendeesCount %d, found %d", attendeesCount, len(attendees))
		return
	}

	john := attendees[0]
	sue := attendees[1]
	travis := attendees[2]

	// check name
	if john.Name != "John Smith" {
		t.Errorf("Expected attendee name %s, found %s", "John Smith", john.Name)
	}
	if sue.Name != "Sue Zimmermann" {
		t.Errorf("Expected attendee name %s, found %s", "Sue Zimmermann", sue.Name)
	}
	if travis.Name != "Travis M. Vollmer" {
		t.Errorf("Expected attendee name %s, found %s", "Travis M. Vollmer", travis.Name)
	}

	// check email
	if john.Email != "j.smith@gmail.com" {
		t.Errorf("Expected attendee email %s, found %s", "j.smith@gmail.com", john.Email)
	}
	if sue.Email != "SueMZimmermann@dayrep.com" {
		t.Errorf("Expected attendee email %s, found %s", "SueMZimmermann@dayrep.com", sue.Email)
	}
	if travis.Email != "travis@dayrep.com" {
		t.Errorf("Expected attendee email %s, found %s", "travis@dayrep.com", travis.Email)
	}

	// check status
	if john.Status != "ACCEPTED" {
		t.Errorf("Expected attendee status %s, found %s", "ACCEPTED", john.Status)
	}
	if sue.Status != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s, found %s", "NEEDS-ACTION", sue.Status)
	}
	if travis.Status != "NEEDS-ACTION" {
		t.Errorf("Expected attendee status %s, found %s", "NEEDS-ACTION", travis.Status)
	}

	// check role
	if john.Role != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", john.Role)
	}
	if sue.Role != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", sue.Role)
	}
	if travis.Role != "REQ-PARTICIPANT" {
		t.Errorf("Expected attendee status %s, found %s", "REQ-PARTICIPANT", travis.Role)
	}
}

func TestCalendarMultidayEvent(t *testing.T) {
	reader := readingFromFile("testCalendars/multiday.ics")
	parser := createParser(reader)
	calendar := newCalendar()
	err := parser.read(calendar)
	parseErrors := parser.getErrors()

	if err != nil {
		t.Errorf("Failed to wait the parse of the calendars ( %s )", err)
	}
	if len(parseErrors) != 0 {
		t.Errorf("Expected 0 error, found %d in :\n  %#v", len(parseErrors), parseErrors)
	}

	// Test a day before the start day
	events := calendar.GetEventIndicesByDate(time.Date(2016, 8, 31, 0, 0, 0, 0, time.UTC))

	// Test exact start day
	events = calendar.GetEventIndicesByDate(time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC))
	if len(events) != 1 {
		t.Errorf("Expected 1 event on the start day, got %d", len(events))
	}

	// Test a random day between start and end date
	events = calendar.GetEventIndicesByDate(time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC))
	if len(events) != 1 {
		t.Errorf("Expected 1 event between start and end, got %d", len(events))
	}

	// Test a day after the end day
	events = calendar.GetEventIndicesByDate(time.Date(2016, 11, 1, 0, 0, 0, 0, time.UTC))
}
