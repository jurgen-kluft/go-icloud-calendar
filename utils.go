package icalendar

import (
	"regexp"
	"strings"
)

// IcsFormat date time format
const IcsFormat = "20060102T150405Z"

// YmdHis Y-m-d H:i:S time format
const YmdHis = "2006-01-02 15:04:05"

// IcsFormatWholeDay ics date format ( describes a whole day)
const IcsFormatWholeDay = "20060102"

func stringToByte(str string) []byte {
	return []byte(str)
}

// removes newlines and cutset from given string
func trimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
}

var mapDayNameToIcsName = map[string]string {
	"Mon": "MO",
	"Tue": "TU",
	"Wed": "WE",
	"Thu": "TH",
	"Fri": "FR",
	"Sat": "ST",
	"Sun": "SU",
}

func getIcsNameFromDayName(day string) string {
	dow, exists := mapDayNameToIcsName[day]
	if !exists {
		dow = ""
	}
	return dow
}
