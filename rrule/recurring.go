package rrule

import (
	"time"
)

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// NextOccurence finds the next occurence of the temporal expression starting at t
func NextOccurence(t time.Time, te TemporalExpression) time.Time {
	t = BeginningOfDay(t)
	for !te.Includes(t) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

// NextN finds the next n occurences of the temportal expression starting at t
func NextN(t time.Time, te TemporalExpression, n int) []time.Time {
	tt := make([]time.Time, n)
	for i := 0; i < n; i++ {
		t = NextOccurence(t, te)
		tt[i] = t
		t = t.Add(24 * time.Hour)
	}
	return tt
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// TemporalExpression matches a subset of time values
type TemporalExpression interface {

	// Include returns true when the provided time matches the temporal expression
	Includes(t time.Time) bool
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// AlwaysOrNeverExpression is a temporal expression that always returns false, it is only here for
// convenience to initialize a TemporalExpression.
type AlwaysOrNeverExpression int

const (
	Never AlwaysOrNeverExpression = iota
	Always
)

// Includes returns true when provided time's day matches the expressions
func (a AlwaysOrNeverExpression) Includes(t time.Time) bool {
	return a == Always
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DayEventExpression is a temporal expression that matches a time-window (start - end) on
// one day.
type DayEventExpression struct {
	Start int
	End   int
}

// Includes returns true when provided time matches the expression
func (t DayEventExpression) Includes(dt time.Time) bool {
	c := dt.Hour()*60 + dt.Minute()
	return c >= t.Start && c < t.End
}

// DayEvents is a helper function that combines multiple DayEventExpression temporal
// expressions with a logical OR operation
func DayEvents(slots ...DayEventExpression) TemporalExpression {
	ee := make([]TemporalExpression, len(slots))
	for i, d := range slots {
		ee[i] = DayEventExpression{Start: d.Start, End: d.End}
	}
	return Or(ee...)
}

// DayEvent is a helper function that creates a single DayEventExpression expression
func DayEvent(start time.Time, end time.Time) TemporalExpression {
	s := start.Hour()*60 + start.Minute()
	e := s + int(end.Sub(start).Minutes())
	slot := DayEventExpression{Start: s, End: e}
	return slot
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// BeforeDateExpression is a temporal expression that matches if a date is
// before a certain set date
type BeforeDateExpression time.Time

// Includes returns true when provided time's day matches the expressions
func (b BeforeDateExpression) Includes(t time.Time) bool {
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return date.Before(time.Time(b))
}

// BeforeDate is a helper function that
func BeforeDate(date time.Time) TemporalExpression {
	return BeforeDateExpression(date)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// AfterDateExpression is a temporal expression that matches if a date is
// after a certain set date
type AfterDateExpression time.Time

// Includes returns true when provided time's day matches the expressions
func (b AfterDateExpression) Includes(t time.Time) bool {
	return t.After(time.Time(b))
}

// AfterDate is a helper function that
func AfterDate(t time.Time, include bool) TemporalExpression {
	date := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999, t.Location())
	if include {
		date = date.AddDate(0, 0, -1)
	}
	return AfterDateExpression(date)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DayExpression is a temporal expression that matches a day of the month starting at 1
// negative numbers start at the end of the month and move backwards
type DayExpression int

func (d DayExpression) normalize(t time.Time) int {
	day := int(d)
	if day < 0 {
		day = EndOfMonth(t).Day() + day + 1
	}
	return day
}

// Includes returns true when provided time's day matches the expressions
func (d DayExpression) Includes(t time.Time) bool {
	return d.normalize(t) == t.Day()
}

// Days is a helper function that combines multiple DayExpression
// objects with a logical OR operation
func Days(days ...int) TemporalExpression {
	ee := make([]TemporalExpression, len(days))
	for i, d := range days {
		ee[i] = DayExpression(d)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DailyExpression is a temporal expression that matches a day using a start date, interval and count
type DailyExpression struct {
	Year     int
	Month    int
	Day      int
	Interval int
	Count    int
}

// Includes returns true when provided date falls in a valid week according to daily
func (t DailyExpression) Includes(c time.Time) bool {
	start := time.Date(t.Year, time.Month(t.Month), t.Day, 0, 0, 0, 0, time.UTC)
	end := time.Date(c.Year(), c.Month(), c.Day(), 0, 0, 0, 0, time.UTC)
	days := int(end.Sub(start).Hours() / 24)
	count := (days + (t.Interval - 1)) / t.Interval
	return ((days % t.Interval) == 0) && ((t.Count == 0) || (count <= t.Count))
}

// Daily is a helper function that creates a single daily expression
func Daily(year int, month int, day int, interval int, count int) TemporalExpression {
	w := DailyExpression{Year: year, Month: month, Day: day, Interval: interval, Count: count}
	return w
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DayRangeExpression is a temporal expression that matches all
// days between the Start and End values
type DayRangeExpression struct {
	Start DayExpression
	End   DayExpression
}

// DayRange returns a temporal expression that matches all
// days between the start and end days
func DayRange(start, end int) DayRangeExpression {
	return DayRangeExpression{DayExpression(start), DayExpression(end)}
}

// Includes returns true when the provided time's day falls
// between the range's Start and Stop values
func (dr DayRangeExpression) Includes(t time.Time) bool {
	d := t.Day()
	return dr.Start.normalize(t) <= d && d <= dr.End.normalize(t)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// WeekInMonthExpression is a temporal expression that matches a week in a month starting at 1
// negative numbers start at the end of the month and move backwards
type WeekInMonthExpression int

func (w WeekInMonthExpression) normalize(t time.Time) int {
	week := int(w)
	if week < 0 {
		week = WeekOfMonth(EndOfMonth(t)) + week + 1
	}
	return week
}

// Includes returns true when the provided time's week matches the expression's
func (w WeekInMonthExpression) Includes(t time.Time) bool {
	return WeekOfMonth(t) == w.normalize(t)
}

// WeeksInMonth is a helper function that combines multiple Week temporal
// expressions with a logical OR operation
func WeeksInMonth(weeks ...int) TemporalExpression {
	ee := make([]TemporalExpression, len(weeks))
	for i, w := range weeks {
		ee[i] = WeekInMonthExpression(w)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// WeeklyExpression is a temporal expression that matches a week using a start date and week interval
type WeeklyExpression struct {
	Year     int
	Month    int
	Day      int
	Interval int
	Count    int
}

// Includes returns true when provided date falls in a valid week according to weekly
func (t WeeklyExpression) Includes(c time.Time) bool {
	start := time.Date(t.Year, time.Month(t.Month), t.Day, 0, 0, 0, 0, time.UTC)
	end := time.Date(c.Year(), c.Month(), c.Day(), 0, 0, 0, 0, time.UTC)
	weeks := int(end.Sub(start).Hours()/24) / 7
	count := (weeks + (t.Interval - 1)) / t.Interval
	return ((weeks % t.Interval) == 0) && ((t.Count == 0) || (count <= t.Count))
}

// Weekly is a helper function that creates a single weekly expression
func Weekly(year int, month int, day int, interval int, count int) TemporalExpression {
	w := WeeklyExpression{Year: year, Month: month, Day: day, Interval: interval, Count: count}
	return w
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// WeekdayExpression is a temporal expression that matches a day of the week
type WeekdayExpression time.Weekday

const (
	Sunday    WeekdayExpression = WeekdayExpression(time.Sunday)
	Monday                      = WeekdayExpression(time.Monday)
	Tuesday                     = WeekdayExpression(time.Tuesday)
	Wednesday                   = WeekdayExpression(time.Wednesday)
	Thursday                    = WeekdayExpression(time.Thursday)
	Friday                      = WeekdayExpression(time.Friday)
	Saturday                    = WeekdayExpression(time.Saturday)
)

// Includes returns true if the provided time's day of the week
// matches the expression's
func (wd WeekdayExpression) Includes(t time.Time) bool {
	return t.Weekday() == time.Weekday(wd)
}

// Weekdays is a helper function that combines multiple Weekday
// temporal expressions using a local OR operation
func Weekdays(weekdays ...time.Weekday) TemporalExpression {
	ee := make([]TemporalExpression, len(weekdays))
	for i, wd := range weekdays {
		ee[i] = WeekdayExpression(wd)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// WeekdayRangeExpression is a temporal expression that matches all
// days between the Start and End values
type WeekdayRangeExpression struct {
	Start time.Weekday
	End   time.Weekday
}

// Includes returns true when the provided time's weekday falls
// between the range's Start and Stop values
func (wr WeekdayRangeExpression) Includes(t time.Time) bool {
	w := t.Weekday()
	return wr.Start <= w && w <= wr.End
}

// WeekdayRange returns a temporal expression that matches all
// days between the start and end values
func WeekdayRange(start, end time.Weekday) WeekdayRangeExpression {
	return WeekdayRangeExpression{start, end}
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DateRangeExpression is a temporal expression that matches all
// days between the Start and End values
type DateRangeExpression struct {
	StartMonth int
	StartDay   int
	EndMonth   int
	EndDay     int
}

// Includes returns true when the provided time's weekday falls
// between the range's Start and Stop values
func (dr DateRangeExpression) Includes(t time.Time) bool {
	m := int(t.Month())
	d := t.Day()
	if m == dr.StartMonth && m == dr.EndMonth {
		return d >= dr.StartDay && d <= dr.EndDay
	} else if m > dr.StartMonth && m < dr.EndMonth {
		return true
	} else if m == dr.StartMonth {
		return d >= dr.StartDay
	} else if m == dr.EndMonth {
		return d <= dr.EndDay
	}
	return false
}

// DateRange returns a temporal expression that matches all
// days between start month:day and end month:day
func DateRange(start, end time.Time) DateRangeExpression {
	return DateRangeExpression{int(start.Month()), start.Day(), int(end.Month()), end.Day()}
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// MonthExpression is a temporal expression which matches a single month
type MonthExpression time.Month

const (
	January   MonthExpression = MonthExpression(time.January)
	February                  = MonthExpression(time.February)
	March                     = MonthExpression(time.March)
	April                     = MonthExpression(time.April)
	May                       = MonthExpression(time.May)
	June                      = MonthExpression(time.June)
	July                      = MonthExpression(time.July)
	August                    = MonthExpression(time.August)
	September                 = MonthExpression(time.September)
	October                   = MonthExpression(time.October)
	November                  = MonthExpression(time.November)
	December                  = MonthExpression(time.December)
)

// Includes returns true when the provided time's date
// matches the temporal expression's
func (m MonthExpression) Includes(t time.Time) bool {
	return t.Month() == time.Month(m)
}

// Months is a helper function that combines multiple Month temporal
// expressions using a local OR operation
func Months(months ...time.Month) TemporalExpression {
	ee := make([]TemporalExpression, len(months))
	for i, m := range months {
		ee[i] = MonthExpression(m)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// MonthlyExpression is a temporal expression that matches a week using a start date and week interval
type MonthlyExpression struct {
	Year     int
	Month    int
	Interval int
	Count    int
}

// Includes returns true when provided date falls in a valid month according to monthly
func (m MonthlyExpression) Includes(c time.Time) bool {
	year := c.Year()
	month := int(c.Month())
	if year < m.Year {
		return false
	} else if year == m.Year && month < m.Month {
		return false
	}
	months := (month + ((year - m.Year) * 12)) - m.Month
	count := months / m.Interval
	return ((months % m.Interval) == 0) && ((m.Count == 0) || (count <= m.Count))
}

// Monthly is a helper function that creates a single monthly expression
func Monthly(year int, month int, interval int, count int) TemporalExpression {
	w := MonthlyExpression{Year: year, Month: month, Interval: interval, Count: count}
	return w
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// MonthRangeExpression is a temporal expression that matches all
// months between the Start and End values
type MonthRangeExpression struct {
	Start time.Month
	End   time.Month
}

// Includes returns true when the provided time's month falls
// between the range's Start and Stop values
func (mr MonthRangeExpression) Includes(t time.Time) bool {
	m := t.Month()
	return mr.Start <= m && m <= mr.End
}

// MonthRange returns a temporal expression that matches all
// months between the start and end values
func MonthRange(start, end time.Month) MonthRangeExpression {
	return MonthRangeExpression{start, end}
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// YearExpression is a temporal expression which matchese a year
type YearExpression int

// Includes returns true when the provided time's year
// matches the temporal expression's
func (y YearExpression) Includes(t time.Time) bool {
	return t.Year() == int(y)
}

// Years is a helper function that combines multipe YearExpression
// objects using a local OR operation
func Years(years ...int) TemporalExpression {
	ee := make([]TemporalExpression, len(years))
	for i, y := range years {
		ee[i] = YearExpression(y)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// YearlyExpression is a temporal expression that matches a year using a start, interval and count
type YearlyExpression struct {
	Year     int
	Interval int
	Count    int
}

// Includes returns true when provided date falls in a valid year according to yearly
func (m YearlyExpression) Includes(c time.Time) bool {
	year := c.Year()
	if year < m.Year {
		return false
	} else if year == m.Year {
		return true
	}
	years := (year - m.Year)
	count := years / m.Interval
	return ((years % m.Interval) == 0) && ((m.Count == 0) || (count <= m.Count))
}

// Yearly is a helper function that creates a single yearly expression
func Yearly(year int, interval int, count int) TemporalExpression {
	w := YearlyExpression{Year: year, Interval: interval, Count: count}
	return w
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// YearRangeExpression is a temporal expression that matches all
// years between the Start and End values
type YearRangeExpression struct {
	Start YearExpression
	End   YearExpression
}

// YearRange returns a temporal expression that matches all
// years between the start and end values
func YearRange(start, end int) YearRangeExpression {
	return YearRangeExpression{YearExpression(start), YearExpression(end)}
}

// Includes returns true when the provided time's years falls
// between the range's Start and Stop values
func (yr YearRangeExpression) Includes(t time.Time) bool {
	year := t.Year()
	return int(yr.Start) <= year && year <= int(yr.End)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// DateExpression is temporal function that matches the year, month, and day
type DateExpression time.Time

// Includes returns true when the provide time's year, month, and
// day match the temporal expression's
func (d DateExpression) Includes(t time.Time) bool {
	y0, m0, d0 := t.Date()
	y1, m1, d1 := time.Time(d).Date()
	return y0 == y1 && m0 == m1 && d0 == d1
}

// Dates is a helper function that combines multiple DateExpression
// objects using a logical OR operation
func Dates(dates ...time.Time) TemporalExpression {
	ee := make([]TemporalExpression, len(dates))
	for i, d := range dates {
		ee[i] = DateExpression(d)
	}
	return Or(ee...)
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// Or combines multiple temporal expressions into one using
// a local Or operation
func Or(ee ...TemporalExpression) OrExpression {
	return OrExpression{ee}
}

// OrExpression is a temporal expression consisting of multiple
// temporal expressions combined using a logical OR operation
type OrExpression struct {
	ee []TemporalExpression
}

// Or adds a temporal expression
func (oe *OrExpression) Or(e TemporalExpression) {
	oe.ee = append(oe.ee, e)
}

// Includes returns true when any of the underlying expressions
// match the provided time
func (oe OrExpression) Includes(t time.Time) bool {
	for _, e := range oe.ee {
		if e.Includes(t) {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// And combines multiple temporal expressions into one using
// a local AND operation
func And(ee ...TemporalExpression) AndExpression {
	return AndExpression{ee}
}

// AndExpression is a temporal expressions consisting of mutliple
// temporal expressions combined with a local AND operation
type AndExpression struct {
	ee []TemporalExpression
}

// And adds a temporal expression
func (ae *AndExpression) And(e TemporalExpression) {
	ae.ee = append(ae.ee, e)
}

// Includes return true when all the underlying temporal expressions
// match the provided time
func (ae AndExpression) Includes(t time.Time) bool {
	for _, e := range ae.ee {
		if !e.Includes(t) {
			return false
		}
	}
	return true
}

// ------------------------------------------------------------------------
// ------------------------------------------------------------------------

// Not negates a temporal expression
func Not(e TemporalExpression) NotExpression {
	return NotExpression{e}
}

// NotExpression is a temporal expression with negates
// its underlying expression
type NotExpression struct {
	e TemporalExpression
}

// Includes returns true when the underlying temporal expression
// does not match the provided time
func (ne NotExpression) Includes(t time.Time) bool {
	return !ne.e.Includes(t)
}
