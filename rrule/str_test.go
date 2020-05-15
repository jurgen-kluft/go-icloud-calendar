package rrule

import (
	"testing"
	"time"
)

func TestRFCRuleToStr(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	dtStart := time.Date(2018, 1, 1, 9, 0, 0, 0, nyLoc)

	r, _ := NewRRule(ROption{Freq: MONTHLY, Dtstart: dtStart, RFC: true})
	if r.String() != "FREQ=MONTHLY" {
		t.Errorf("Expected RFC string FREQ=MONTHLY, got %v", r.String())
	}
}

func TestRuleToStr(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	dtStart := time.Date(2018, 1, 1, 9, 0, 0, 0, nyLoc)

	r, _ := NewRRule(ROption{Freq: MONTHLY, Dtstart: dtStart})
	if r.String() != "FREQ=MONTHLY;DTSTART=20180101T140000Z" {
		t.Errorf("Expected non RFC string FREQ=MONTHLY;DTSTART=20180101T140000Z, got %v", r.String())
	}
}

func TestStr(t *testing.T) {
	str := "FREQ=WEEKLY;DTSTART=20120201T093000Z;INTERVAL=5;WKST=TU;COUNT=2;UNTIL=20130130T230000Z;BYSETPOS=2;BYMONTH=3;BYYEARDAY=95;BYWEEKNO=1;BYDAY=MO,+2FR;BYHOUR=9;BYMINUTE=30;BYSECOND=0;BYEASTER=-1"
	r, _ := StrToRRule(str)
	if s := r.String(); s != str {
		t.Errorf("StrToRRule(%q).String() = %q, want %q", str, s, str)
	}

	if r.OrigOptions.RFC {
		t.Errorf("StrToRRule(%q).OrigOptions.RFC = true, want false", str)
	}
}

func TestInvalidString(t *testing.T) {
	cases := []string{
		"",
		"    ",
		"FREQ",
		"FREQ=HELLO",
		"BYMONTH=",
		"FREQ=WEEKLY;HELLO=WORLD",
		"FREQ=WEEKLY;BYMONTHDAY=I",
		"FREQ=WEEKLY;BYDAY=M",
		"FREQ=WEEKLY;BYDAY=MQ",
		"FREQ=WEEKLY;BYDAY=+MO",
		"BYDAY=MO",
	}
	for _, item := range cases {
		if _, e := StrToRRule(item); e == nil {
			t.Errorf("StrToRRule(%q) = nil, want error", item)
		}
	}
}

func TestStrToDtStart(t *testing.T) {
	validCases := []string{
		"19970714T133000",
		"19970714T173000Z",
		"TZID=America/New_York:19970714T133000",
	}

	invalidCases := []string{
		"DTSTART;TZID=America/New_York:19970714T133000",
		"19970714T1330000",
		"DTSTART;TZID=:20180101T090000",
		"TZID=:20180101T090000",
		"TZID=notatimezone:20180101T090000",
		"DTSTART:19970714T133000",
		"DTSTART:19970714T133000Z",
		"DTSTART;:19970714T133000Z",
		"DTSTART;:1997:07:14T13:30:00Z",
		";:19970714T133000Z",
		"    ",
		"",
	}

	for _, item := range validCases {
		if _, e := strToDtStart(item, time.UTC); e != nil {
			t.Errorf("strToDtStart(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := strToDtStart(item, time.UTC); e == nil {
			t.Errorf("strToDtStart(%q) err = nil, want not nil", item)
		}
	}
}

func TestStrToDates(t *testing.T) {
	validCases := []string{
		"19970714T133000",
		"19970714T173000Z",
		"VALUE=DATE-TIME:19970714T133000,19980714T133000,19980714T133000",
		"VALUE=DATE-TIME;TZID=America/New_York:19970714T133000,19980714T133000,19980714T133000",
		"VALUE=DATE:19970714T133000,19980714T133000,19980714T133000",
	}

	invalidCases := []string{
		"VALUE:DATE:TIME:19970714T133000,19980714T133000,19980714T133000",
		";:19970714T133000Z",
		"    ",
		"",
		"VALUE=DATE-TIME;TZID=:19970714T133000",
		"VALUE=PERIOD:19970714T133000Z/19980714T133000Z",
	}

	for _, item := range validCases {
		if _, e := StrToDates(item); e != nil {
			t.Errorf("StrToDates(%q) error = %s, want nil", item, e.Error())
		}
		if _, e := StrToDatesInLoc(item, time.Local); e != nil {
			t.Errorf("StrToDates(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := StrToDates(item); e == nil {
			t.Errorf("StrToDates(%q) err = nil, want not nil", item)
		}
		if _, e := StrToDatesInLoc(item, time.Local); e == nil {
			t.Errorf("StrToDates(%q) err = nil, want not nil", item)
		}
	}
}

func TestStrToDatesTimeIsCorrect(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	inputs := []string{
		"VALUE=DATE-TIME:19970714T133000",
		"VALUE=DATE-TIME;TZID=America/New_York:19970714T133000",
	}
	exp := []time.Time{
		time.Date(1997, 7, 14, 13, 30, 0, 0, time.UTC),
		time.Date(1997, 7, 14, 13, 30, 0, 0, nyLoc),
	}

	for i, s := range inputs {
		ts, err := StrToDates(s)
		if err != nil {
			t.Fatalf("StrToDates(%s): error = %s", s, err.Error())
		}
		if len(ts) != 1 {
			t.Fatalf("StrToDates(%s): bad answer: %v", s, ts)
		}
		if !ts[0].Equal(exp[i]) {
			t.Fatalf("StrToDates(%s): bad answer: %v, expected: %v", s, ts[0], exp[i])
		}
	}
}

func TestProcessRRuleName(t *testing.T) {
	validCases := []string{
		"DTSTART;TZID=America/New_York:19970714T133000",
		"RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU",
		"EXRULE:FREQ=WEEKLY;INTERVAL=4;BYDAY=MO",
		"EXDATE;VALUE=DATE-TIME:20180525T070000Z,20180530T130000Z",
		"RDATE;TZID=America/New_York;VALUE=DATE-TIME:20180801T131313Z,20180902T141414Z",
	}

	invalidCases := []string{
		"TZID=America/New_York:19970714T133000",
		"19970714T1330000",
		";:19970714T133000Z",
		"FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU",
		"    ",
	}

	for _, item := range validCases {
		if _, e := processRRuleName(item); e != nil {
			t.Errorf("processRRuleName(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := processRRuleName(item); e == nil {
			t.Errorf("processRRuleName(%q) err = nil, want not nil", item)
		}
	}
}
