package rrule

import (
	"errors"
	"fmt"
	"time"
)

// Every mask is 7 days longer to handle cross-year weekly periods.
var (
	M366MASK     []int
	M365MASK     []int
	MDAY366MASK  []int
	MDAY365MASK  []int
	NMDAY366MASK []int
	NMDAY365MASK []int
	WDAYMASK     []int
	M366RANGE    = []int{0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335, 366}
	M365RANGE    = []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365}
)

func init() {
	M366MASK = concat(repeat(1, 31), repeat(2, 29), repeat(3, 31),
		repeat(4, 30), repeat(5, 31), repeat(6, 30), repeat(7, 31),
		repeat(8, 31), repeat(9, 30), repeat(10, 31), repeat(11, 30),
		repeat(12, 31), repeat(1, 7))
	M365MASK = concat(M366MASK[:59], M366MASK[60:])
	M29, M30, M31 := rang(1, 30), rang(1, 31), rang(1, 32)
	MDAY366MASK = concat(M31, M29, M31, M30, M31, M30, M31, M31, M30, M31, M30, M31, M31[:7])
	MDAY365MASK = concat(MDAY366MASK[:59], MDAY366MASK[60:])
	M29, M30, M31 = rang(-29, 0), rang(-30, 0), rang(-31, 0)
	NMDAY366MASK = concat(M31, M29, M31, M30, M31, M30, M31, M31, M30, M31, M30, M31, M31[:7])
	NMDAY365MASK = concat(NMDAY366MASK[:31], NMDAY366MASK[32:])
	for i := 0; i < 55; i++ {
		WDAYMASK = append(WDAYMASK, []int{0, 1, 2, 3, 4, 5, 6}...)
	}
}

// Frequency denotes the period on which the rule is evaluated.
type Frequency int

// Constants
const (
	YEARLY Frequency = iota
	MONTHLY
	WEEKLY
	DAILY
	HOURLY
	MINUTELY
	SECONDLY
)

// RWeekday specifying the nth weekday.
// Field N could be positive or negative (like MO(+2) or MO(-3).
// Not specifying N (0) is the same as specifying +1.
type RWeekday struct {
	weekday int
	n       int
}

// Nth return the nth weekday
// __call__ - Cannot call the object directly,
// do it through e.g. TH.nth(-1) instead,
func (wday *RWeekday) Nth(n int) RWeekday {
	return RWeekday{wday.weekday, n}
}

// N returns index of the week, e.g. for 3MO, N() will return 3
func (wday *RWeekday) N() int {
	return wday.n
}

// Day returns index of the day in a week (0 for MO, 6 for SU)
func (wday *RWeekday) Day() int {
	return wday.weekday
}

// Weekdays
var (
	MO = RWeekday{weekday: 0}
	TU = RWeekday{weekday: 1}
	WE = RWeekday{weekday: 2}
	TH = RWeekday{weekday: 3}
	FR = RWeekday{weekday: 4}
	SA = RWeekday{weekday: 5}
	SU = RWeekday{weekday: 6}
)

// ROption offers options to construct a RRule instance
type ROption struct {
	Code       string
	Freq       Frequency
	Dtstart    time.Time
	Interval   int
	Wkst       RWeekday
	Count      int
	Until      time.Time
	Bysetpos   []int
	Bymonth    []int
	Bymonthday []int
	Byyearday  []int
	Byweekno   []int
	Byweekday  []RWeekday
	Byhour     []int
	Byminute   []int
	Bysecond   []int
	Byeaster   []int
	RFC        bool
}

// RRule offers a small, complete, and very fast, implementation of the recurrence rules
// documented in the iCalendar RFC, including support for caching of results.
type RRule struct {
	OrigOptions             ROption
	Options                 ROption
	freq                    Frequency
	dtstart                 time.Time
	interval                int
	wkst                    int
	count                   int
	until                   time.Time
	bysetpos                []int
	bymonth                 []int
	bymonthday, bynmonthday []int
	byyearday               []int
	byweekno                []int
	byweekday               []int
	bynweekday              []RWeekday
	byhour                  []int
	byminute                []int
	bysecond                []int
	byeaster                []int
	len                     int
	compiled                TemporalExpression
}

// NOTE
// Try to find a way to convert most of the RRule configuration to a 'recurring' TemporalExpression.
// If we can then we can remove a lot of the python code (Iterator etc..) and should be a lot more
// efficient.

// NewRRule construct a new RRule instance
func NewRRule(arg ROption) (*RRule, error) {
	if err := validateBounds(arg); err != nil {
		return nil, err
	}
	r := RRule{}
	r.OrigOptions = arg
	if arg.Dtstart.IsZero() {
		arg.Dtstart = time.Now().UTC()
	}
	arg.Dtstart = arg.Dtstart.Truncate(time.Second)
	r.dtstart = arg.Dtstart
	r.freq = arg.Freq
	if arg.Interval == 0 {
		r.interval = 1
	} else {
		r.interval = arg.Interval
	}
	r.count = arg.Count
	if arg.Until.IsZero() {
		// add largest representable duration (approximately 290 years).
		arg.Until = r.dtstart.Add(time.Duration(1<<63 - 1))
	}
	r.until = arg.Until
	r.wkst = arg.Wkst.weekday
	r.bysetpos = arg.Bysetpos
	if len(arg.Byweekno) == 0 &&
		len(arg.Byyearday) == 0 &&
		len(arg.Bymonthday) == 0 &&
		len(arg.Byweekday) == 0 &&
		len(arg.Byeaster) == 0 {
		if r.freq == YEARLY {
			if len(arg.Bymonth) == 0 {
				arg.Bymonth = []int{int(r.dtstart.Month())}
			}
			arg.Bymonthday = []int{r.dtstart.Day()}
		} else if r.freq == MONTHLY {
			arg.Bymonthday = []int{r.dtstart.Day()}
		} else if r.freq == WEEKLY {
			arg.Byweekday = []RWeekday{RWeekday{weekday: toPyWeekday(r.dtstart.Weekday())}}
		}
	}
	r.bymonth = arg.Bymonth
	r.byyearday = arg.Byyearday
	r.byeaster = arg.Byeaster
	for _, mday := range arg.Bymonthday {
		if mday > 0 {
			r.bymonthday = append(r.bymonthday, mday)
		} else if mday < 0 {
			r.bynmonthday = append(r.bynmonthday, mday)
		}
	}
	r.byweekno = arg.Byweekno
	for _, wday := range arg.Byweekday {
		if wday.n == 0 || r.freq > MONTHLY {
			r.byweekday = append(r.byweekday, wday.weekday)
		} else {
			r.bynweekday = append(r.bynweekday, wday)
		}
	}
	if len(arg.Byhour) == 0 {
		if r.freq < HOURLY {
			r.byhour = []int{r.dtstart.Hour()}
		}
	} else {
		r.byhour = arg.Byhour
	}
	if len(arg.Byminute) == 0 {
		if r.freq < MINUTELY {
			r.byminute = []int{r.dtstart.Minute()}
		}
	} else {
		r.byminute = arg.Byminute
	}
	if len(arg.Bysecond) == 0 {
		if r.freq < SECONDLY {
			r.bysecond = []int{r.dtstart.Second()}
		}
	} else {
		r.bysecond = arg.Bysecond
	}

	r.Options = arg
	r.compiled = Never

	return &r, nil
}

// validateBounds checks the RRule's options are within the boundaries defined
// in RRFC 5545. This is useful to ensure that the RRule can even have any times,
// as going outside these bounds trivially will never have any dates. This can catch
// obvious user error.
func validateBounds(arg ROption) error {
	bounds := []struct {
		field     []int
		param     string
		bound     []int
		plusMinus bool // If the bound also applies for -x to -y.
	}{
		{arg.Bysecond, "bysecond", []int{0, 59}, false},
		{arg.Byminute, "byminute", []int{0, 59}, false},
		{arg.Byhour, "byhour", []int{0, 23}, false},
		{arg.Bymonthday, "bymonthday", []int{1, 31}, true},
		{arg.Byyearday, "byyearday", []int{1, 366}, true},
		{arg.Byweekno, "byweekno", []int{1, 53}, true},
		{arg.Bymonth, "bymonth", []int{1, 12}, false},
		{arg.Bysetpos, "bysetpos", []int{1, 366}, true},
	}

	checkBounds := func(param string, value int, bounds []int, plusMinus bool) error {
		if !(value >= bounds[0] && value <= bounds[1]) && (!plusMinus || !(value <= -bounds[0] && value >= -bounds[1])) {
			plusMinusBounds := ""
			if plusMinus {
				plusMinusBounds = fmt.Sprintf(" or %d and %d", -bounds[0], -bounds[1])
			}
			return fmt.Errorf("%s must be between %d and %d%s", param, bounds[0], bounds[1], plusMinusBounds)
		}
		return nil
	}

	for _, b := range bounds {
		for _, value := range b.field {
			if err := checkBounds(b.param, value, b.bound, b.plusMinus); err != nil {
				return err
			}
		}
	}

	// Days can optionally specify weeks, like BYDAY=+2MO for the 2nd Monday
	// of the month/year.
	for _, w := range arg.Byweekday {
		if w.n > 53 || w.n < -53 {
			return errors.New("byday must be between 1 and 53 or -1 and -53")
		}
	}

	if arg.Interval < 0 {
		return errors.New("interval must be greater than 0")
	}

	return nil
}

// Just to satisfy the test suite
func (r *RRule) All() []time.Time {
	return []time.Time{}
}
func (r *RRule) Until(dt time.Time) []time.Time {
	return []time.Time{}
}
func (r *RRule) Before(dt time.Time, include bool) time.Time {
	return time.Time{}
}
func (r *RRule) After(dt time.Time, include bool) time.Time {
	return time.Time{}
}
func (r *RRule) Between(dts time.Time, dte time.Time, include bool) []time.Time {
	return []time.Time{}
}

func (r *RRule) DTStart(dt time.Time) {
	r.dtstart = dt.Truncate(time.Second)
	r.Options.Dtstart = r.dtstart

	if len(r.Options.Byhour) == 0 && r.freq < HOURLY {
		r.byhour = []int{r.dtstart.Hour()}
	}
	if len(r.Options.Byminute) == 0 && r.freq < MINUTELY {
		r.byminute = []int{r.dtstart.Minute()}
	}
	if len(r.Options.Bysecond) == 0 && r.freq < SECONDLY {
		r.bysecond = []int{r.dtstart.Second()}
	}
}

// Compile will convert the RRule information into a TemporalExpression structure
func (r *RRule) Compile(start time.Time, end time.Time) error {
	if r.freq == YEARLY {
		r.compiled = And(AfterDate(start, true), Yearly(start.Year(), r.interval, r.count), DateRange(start, end))
		return nil
	} else if r.freq == MONTHLY {
		r.compiled = And(AfterDate(start, true), Monthly(start.Year(), int(start.Month()), r.interval, r.count), BeforeDate(end))
		return nil
	} else if r.freq == DAILY {
		r.compiled = And(AfterDate(start, true), Daily(start.Year(), int(start.Month()), start.Day(), r.interval, r.count), BeforeDate(end))
		return nil
	}

	r.compiled = Never
	return errors.New("Failed to compile RRule '%s' into temporal expression")
}

// Includes will determine if 'dt' is regarded by the TemporalExpression as included
func (r *RRule) Includes(dt time.Time) bool {
	return r.compiled.Includes(dt)
}
