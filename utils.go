package icalendar

import (
	"os"
	"regexp"
	"strings"
)

// RepeatRuleApply is true , the rrule will create new objects for the repeated events
var RepeatRuleApply bool

// MaxRepeats max of the rrule repeat for single event
var MaxRepeats int

//  unixtimestamp
const uts = "1136239445"

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

//  checks if file exists
func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func parseDayNameToIcsName(day string) string {
	var dow string
	switch day {
	case "Mon":
		dow = "MO"
		break
	case "Tue":
		dow = "TU"
		break
	case "Wed":
		dow = "WE"
		break
	case "Thu":
		dow = "TH"
		break
	case "Fri":
		dow = "FR"
		break
	case "Sat":
		dow = "ST"
		break
	case "Sun":
		dow = "SU"
		break
	default:
		// fmt.Println("DEFAULT :", start.Format("Mon"))
		dow = ""
		break
	}
	return dow
}
