package icalendar

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type parser struct {
	reader          Reader
	repeatRuleApply bool // RepeatRuleApply is true , the rrule will create new objects for the repeated events
	maxRepeats      int  // MaxRepeats max of the rrule repeat for single event
	errorsOccured   []error
	parsedEvents    []*Event
}

// creates new parser
func createParser(r Reader) *parser {
	p := new(parser)
	p.reader = r
	p.repeatRuleApply = false
	p.maxRepeats = 10
	p.errorsOccured = []error{}
	p.parsedEvents = []*Event{}
	return p
}

func (p *parser) reset() {
	p.errorsOccured = []error{}
	p.parsedEvents = []*Event{}
}

func (p *parser) read(cal *Calendar) error {
	p.reset()

	content, err := p.reader.Read()
	if err != nil {
		p.errorsOccured = append(p.errorsOccured, err)
		return err
	}

	p.parseContent(cal, content)

	for _, e := range p.errorsOccured {
		fmt.Println(e)
	}
	if len(p.errorsOccured) == 0 {
		return nil
	}
	return fmt.Errorf("Reading calendar content has errors")
}

func (p *parser) getErrors() []error {
	return p.errorsOccured
}

// PARSING

func (p *parser) parseContent(ical *Calendar, content string) {
	// split the data into calendar info and events data
	eventsData, calInfo := explodeICal(content)

	// set the calendar properties
	ical.Name = (p.parseICalName(calInfo))
	ical.Description = (p.parseICalDesc(calInfo))
	ical.Version = (p.parseICalVersion(calInfo))
	ical.Timezone = (p.parseICalTimezone(calInfo))

	// parse all events and add them to the calendar
	p.parseEvents(ical, eventsData)
}

func explodeICal(content string) ([]string, string) {
	reEvents, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	allEvents := reEvents.FindAllString(content, len(content))
	calInfo := reEvents.ReplaceAllString(content, "")
	return allEvents, calInfo
}

func (p *parser) parseICalName(content string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.FindString(content)
	return trimField(result, "X-WR-CALNAME:")
}

func (p *parser) parseICalDesc(content string) string {
	re, _ := regexp.Compile(`X-WR-CALDESC:.*?\n`)
	result := re.FindString(content)
	return trimField(result, "X-WR-CALDESC:")
}

func (p *parser) parseICalVersion(content string) float64 {
	re, _ := regexp.Compile(`VERSION:.*?\n`)
	result := re.FindString(content)
	// parse the version result to float
	ver, _ := strconv.ParseFloat(trimField(result, "VERSION:"), 64)
	return ver
}

func (p *parser) parseICalTimezone(content string) *time.Location {
	re, _ := regexp.Compile(`X-WR-TIMEZONE:.*?\n`)
	result := re.FindString(content)

	// parse the timezone result to time.Location
	timezone := trimField(result, "X-WR-TIMEZONE:")
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		p.errorsOccured = append(p.errorsOccured, err)
		loc, _ = time.LoadLocation("UTC")
	}
	return loc
}

// EVENTS PARSING

func (p *parser) parseEvents(cal *Calendar, eventsData []string) {
	for _, eventData := range eventsData {
		event := NewEvent()

		start := p.parseEventStart(eventData)
		end := p.parseEventEnd(eventData)
		// whole day event when both times are 00:00:00
		wholeDay := start.Hour() == 0 && end.Hour() == 0 && start.Minute() == 0 && end.Minute() == 0 && start.Second() == 0 && end.Second() == 0

		event.Status = (p.parseEventStatus(eventData))
		event.Summary = (p.parseEventSummary(eventData))
		event.Description = (p.parseEventDescription(eventData))
		event.ImportedID = (p.parseEventID(eventData))
		event.Class = (p.parseEventClass(eventData))
		event.Sequence = (p.parseEventSequence(eventData))
		event.Created = (p.parseEventCreated(eventData))
		event.Modified = (p.parseEventModified(eventData))
		event.Rrule = (p.parseEventRRule(eventData))
		event.Location = (p.parseEventLocation(eventData))
		event.Geo = (p.parseEventGeo(eventData))
		event.Start = (start)
		event.End = (end)
		event.IsWholeDayEvent = (wholeDay)
		event.Attendees = (p.parseEventAttendees(eventData))
		event.Organizer = (p.parseEventOrganizer(eventData))
		event.Owner = (cal)
		event.ID = (event.GenerateUUID())

		err := cal.InsertEvent(event)
		if err != nil {
			p.errorsOccured = append(p.errorsOccured, err)
		}
	}
}

func (p *parser) parseEventSummary(eventData string) string {
	re, _ := regexp.Compile(`SUMMARY:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "SUMMARY:")
}

func (p *parser) parseEventStatus(eventData string) string {
	re, _ := regexp.Compile(`STATUS:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "STATUS:")
}

func (p *parser) parseEventDescription(eventData string) string {
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)
	result := re.FindString(eventData)
	return trimField(strings.Replace(result, "\r\n ", "", -1), "DESCRIPTION:")
}

func (p *parser) parseEventID(eventData string) string {
	re, _ := regexp.Compile(`UID:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "UID:")
}

func (p *parser) parseEventClass(eventData string) string {
	re, _ := regexp.Compile(`CLASS:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "CLASS:")
}

func (p *parser) parseEventSequence(eventData string) int {
	re, _ := regexp.Compile(`SEQUENCE:.*?\n`)
	result := re.FindString(eventData)
	sq, _ := strconv.Atoi(trimField(result, "SEQUENCE:"))
	return sq
}

func (p *parser) parseEventCreated(eventData string) time.Time {
	re, _ := regexp.Compile(`CREATED:.*?\n`)
	result := re.FindString(eventData)
	created := trimField(result, "CREATED:")
	t, _ := time.Parse(IcsFormat, created)
	return t
}

func (p *parser) parseEventModified(eventData string) time.Time {
	re, _ := regexp.Compile(`LAST-MODIFIED:.*?\n`)
	result := re.FindString(eventData)
	modified := trimField(result, "LAST-MODIFIED:")
	t, _ := time.Parse(IcsFormat, modified)
	return t
}

func (p *parser) parseEventStart(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTSTART;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTSTART(;TZID=.*?){0,1}:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTSTART;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTSTART(;TZID=.*?){0,1}:")

		if !strings.Contains(modified, "Z") {
			modified = fmt.Sprintf("%sZ", modified)
		}

		t, _ = time.Parse(IcsFormat, modified)
	}

	return t
}

func (p *parser) parseEventEnd(eventData string) time.Time {
	reWholeDay, _ := regexp.Compile(`DTEND;VALUE=DATE:.*?\n`)
	re, _ := regexp.Compile(`DTEND(;TZID=.*?){0,1}:.*?\n`)
	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, "DTEND;VALUE=DATE:")
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has end hour and minute
		result := re.FindString(eventData)
		modified := trimField(result, "DTEND(;TZID=.*?){0,1}:")

		if !strings.Contains(modified, "Z") {
			modified = fmt.Sprintf("%sZ", modified)
		}
		t, _ = time.Parse(IcsFormat, modified)
	}
	return t

}

func (p *parser) parseEventRRule(eventData string) string {
	re, _ := regexp.Compile(`RRULE:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "RRULE:")
}

func (p *parser) parseEventLocation(eventData string) string {
	re, _ := regexp.Compile(`LOCATION:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "LOCATION:")
}

func (p *parser) parseEventGeo(eventData string) *Geo {
	re, _ := regexp.Compile(`GEO:.*?\n`)
	result := re.FindString(eventData)

	value := trimField(result, "GEO:")
	values := strings.Split(value, ";")
	if len(values) < 2 {
		return nil
	}

	return NewGeo(values[0], values[1])
}

// ATTENDEE PARSING

func (p *parser) parseEventAttendees(eventData string) []*Attendee {
	attendeesObj := []*Attendee{}
	re, _ := regexp.Compile(`ATTENDEE(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	attendees := re.FindAllString(eventData, len(eventData))

	for _, attendeeData := range attendees {
		if attendeeData == "" {
			continue
		}
		attendee := p.parseAttendee(strings.Replace(strings.Replace(attendeeData, "\r", "", 1), "\n ", "", 1))
		//  check for any fields set
		if attendee.Email != "" || attendee.Name != "" || attendee.Role != "" || attendee.Status != "" || attendee.Type != "" {
			attendeesObj = append(attendeesObj, attendee)
		}
	}
	return attendeesObj
}

func (p *parser) parseEventOrganizer(eventData string) *Attendee {

	re, _ := regexp.Compile(`ORGANIZER(:|;)(.*?\r?\n)(\s.*?\r?\n)*`)
	organizerData := re.FindString(eventData)
	if organizerData == "" {
		return nil
	}
	organizerDataFormated := strings.Replace(strings.Replace(organizerData, "\r", "", 1), "\n ", "", 1)

	a := NewAttendee()
	a.Email = (p.parseAttendeeMail(organizerDataFormated))
	a.Name = (p.parseOrganizerName(organizerDataFormated))

	return a
}

func (p *parser) parseAttendee(attendeeData string) *Attendee {

	a := NewAttendee()
	a.Email = (p.parseAttendeeMail(attendeeData))
	a.Name = (p.parseAttendeeName(attendeeData))
	a.Role = (p.parseAttendeeRole(attendeeData))
	a.Status = (p.parseAttendeeStatus(attendeeData))
	a.Type = (p.parseAttendeeType(attendeeData))
	return a
}

func (p *parser) parseAttendeeMail(attendeeData string) string {
	re, _ := regexp.Compile(`mailto:.*?\n`)
	result := re.FindString(attendeeData)
	return trimField(result, "mailto:")
}

func (p *parser) parseAttendeeStatus(attendeeData string) string {
	re, _ := regexp.Compile(`PARTSTAT=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(PARTSTAT=|;)`)
}

func (p *parser) parseAttendeeRole(attendeeData string) string {
	re, _ := regexp.Compile(`ROLE=.*?;`)
	result := re.FindString(attendeeData)

	if result == "" {
		return ""
	}
	return trimField(result, `(ROLE=|;)`)
}

func (p *parser) parseAttendeeName(attendeeData string) string {
	re, _ := regexp.Compile(`CN=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|;)`)
}

// parses the organizer Name
func (p *parser) parseOrganizerName(orgData string) string {
	re, _ := regexp.Compile(`CN=.*?:`)
	result := re.FindString(orgData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CN=|:)`)
}

func (p *parser) parseAttendeeType(attendeeData string) string {
	re, _ := regexp.Compile(`CUTYPE=.*?;`)
	result := re.FindString(attendeeData)
	if result == "" {
		return ""
	}
	return trimField(result, `(CUTYPE=|;)`)
}
