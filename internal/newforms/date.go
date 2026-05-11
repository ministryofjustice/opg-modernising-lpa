package newforms

import (
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type Date struct {
	Field      // Input not set
	Inputs     []string
	Value      date.Date
	validators []func(date.Date) Error
}

func NewDate(name, label string) *Date {
	f := &Date{}
	f.Name = name
	f.Label = label
	return f
}

func (f *Date) SetInput(d date.Date) {
	f.Inputs = []string{d.YearString(), d.MonthString(), d.DayString()}
}

func (f *Date) NotEmpty() *Date {
	f.validators = append(f.validators, func(d date.Date) Error {
		e := DateMissingError{Field: f.Field}

		if d.DayString() == "" {
			e.MissingDay = true
		}
		if d.MonthString() == "" {
			e.MissingMonth = true
		}
		if d.YearString() == "" {
			e.MissingYear = true
		}

		if e.MissingDay || e.MissingMonth || e.MissingYear {
			if e.MissingDay && e.MissingMonth && e.MissingYear {
				return EmptyError{Field: f.Field}
			}

			return e
		}

		return nil
	})

	return f
}

func (f *Date) MustBeReal() *Date {
	f.validators = append(f.validators, func(d date.Date) Error {
		if !d.Valid() {
			return DateMustBeRealError{Field: f.Field}
		}

		return nil
	})

	return f
}

func (f *Date) MustBePast() *Date {
	f.validators = append(f.validators, func(d date.Date) Error {
		if d.After(date.Today()) {
			return DateMustBePastError{Field: f.Field}
		}

		return nil
	})

	return f
}

func (f *Date) BeforeYears(years int, errorMessage string) *Date {
	f.validators = append(f.validators, func(d date.Date) Error {
		if d.After(date.Today().AddDate(-years, 0, 0)) {
			return CustomError(errorMessage)
		}

		return nil
	})

	return f
}

func (f *Date) Parse(values url.Values) {
	f.Inputs = []string{
		strings.TrimSpace(values.Get(f.Name + "-year")),
		strings.TrimSpace(values.Get(f.Name + "-month")),
		strings.TrimSpace(values.Get(f.Name + "-day")),
	}
	f.Value = date.New(f.Inputs[0], f.Inputs[1], f.Inputs[2])

	for _, validator := range f.validators {
		if error := validator(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

func (f *Date) ErrorPart(part string) bool {
	if err, ok := f.Error.(DateMissingError); ok {
		switch part {
		case "day":
			return err.MissingDay
		case "month":
			return err.MissingMonth
		case "year":
			return err.MissingYear
		}
	}

	return false
}
