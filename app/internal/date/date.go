package date

import (
	"fmt"
	"strings"
	"time"
)

const dateFormat = "2006-1-2"

type FormatError string

func (e FormatError) Error() string {
	return fmt.Sprintf("date '%s' incorrectly formatted", string(e))
}

type Date struct {
	year  string
	month string
	day   string

	t   time.Time
	err error
}

func New(year, month, day string) Date {
	t, err := time.Parse(dateFormat, year+"-"+month+"-"+day)

	return Date{
		year:  year,
		month: month,
		day:   day,
		t:     t,
		err:   err,
	}
}

func Today() Date {
	return fromTime(time.Now().UTC().Round(24 * time.Hour))
}

func fromTime(t time.Time) Date {
	return Date{
		year:  t.Format("2006"),
		month: t.Format("1"),
		day:   t.Format("2"),
		t:     t,
	}
}

func (d Date) Year() string {
	return d.year
}

func (d Date) Month() string {
	return d.month
}

func (d Date) Day() string {
	return d.day
}

func (d Date) Valid() bool {
	return d.err == nil
}

func (d Date) IsZero() bool {
	return d.t.IsZero()
}

func (d Date) String() string {
	return d.t.Format(dateFormat)
}

func (d Date) Before(other Date) bool {
	return d.t.Before(other.t)
}

func (d Date) After(other Date) bool {
	return d.t.After(other.t)
}

func (d Date) AddDate(years, months, days int) Date {
	return fromTime(d.t.AddDate(years, months, days))
}

func (d *Date) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	datePart, _, _ := strings.Cut(string(text), "T")

	parts := strings.Split(datePart, "-")
	if len(parts) != 3 {
		return FormatError(text)
	}

	*d = New(parts[0], parts[1], parts[2])

	return nil
}

func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return []byte{}, nil
	}

	return []byte(d.String()), nil
}
