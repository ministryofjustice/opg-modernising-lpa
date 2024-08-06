// Package date provides functionality for working with dates.
package date

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const unpaddedDate = "2006-1-2"

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

type TimeOrDate interface {
	IsZero() bool
	Format(string) string
	Day() int
	Month() time.Month
	Year() int
}

func New(year, month, day string) Date {
	t, err := time.Parse(unpaddedDate, year+"-"+month+"-"+day)

	return Date{
		year:  year,
		month: month,
		day:   day,
		t:     t,
		err:   err,
	}
}

func Today() Date {
	return FromTime(time.Now())
}

func FromTime(t time.Time) Date {
	if t.IsZero() {
		return Date{}
	}

	return Date{
		year:  t.Format("2006"),
		month: t.Format("01"),
		day:   t.Format("02"),
		t:     t.Truncate(24 * time.Hour),
	}
}

func (d Date) Year() int {
	return d.t.Year()
}

func (d Date) Month() time.Month {
	return d.t.Month()
}

func (d Date) Day() int {
	return d.t.Day()
}

func (d Date) YearString() string {
	return d.year
}

func (d Date) MonthString() string {
	return d.month
}

func (d Date) DayString() string {
	return d.day
}

func (d Date) Valid() bool {
	return d.err == nil
}

func (d Date) IsZero() bool {
	return d.t.IsZero()
}

func (d Date) Format(format string) string {
	return d.t.Format(format)
}

func (d Date) String() string {
	return d.t.Format(unpaddedDate)
}

func (d Date) Equals(other Date) bool {
	return d.String() == other.String()
}

func (d Date) Before(other Date) bool {
	return d.t.Before(other.t)
}

func (d Date) After(other Date) bool {
	return d.t.After(other.t)
}

func (d Date) AddDate(years, months, days int) Date {
	return FromTime(d.t.AddDate(years, months, days))
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

	return []byte(d.t.Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return d.UnmarshalText([]byte(s))
}

func (d Date) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	text := ""
	if !d.IsZero() {
		text = d.t.Format(unpaddedDate)
	}

	return attributevalue.Marshal(text)
}

func (d Date) Time() time.Time {
	return d.t
}

func (d Date) Hash() (uint64, error) {
	return uint64(d.t.Unix()), nil
}
