package forms

import (
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

// Date is a validatable date form field.
type Date struct {
	Field                           // Input is unused
	InputDay, InputMonth, InputYear string
	Value                           date.Date
	validators                      []validator[date.Date]
}

func NewDate(name, label string) *Date {
	f := &Date{}
	f.Name = name
	f.Label = label
	return f
}

func (f *Date) Set(d date.Date) {
	f.InputDay = d.DayString()
	f.InputMonth = d.MonthString()
	f.InputYear = d.YearString()
}

func (f *Date) Parse(values url.Values) {
	f.InputYear = strings.TrimSpace(values.Get(f.Name + "-year"))
	f.InputMonth = strings.TrimSpace(values.Get(f.Name + "-month"))
	f.InputDay = strings.TrimSpace(values.Get(f.Name + "-day"))
	f.Value = date.New(f.InputYear, f.InputMonth, f.InputDay)

	for _, validator := range f.validators {
		if error := validator.Validate(f.Value); error != nil {
			f.Error = error
			break
		}
	}
}

// WithError overrides the error returned by the previously defined validator.
func (f *Date) WithError(replace Error) *Date {
	if l := len(f.validators); l > 0 {
		f.validators[l-1] = withError[date.Date]{replace: replace, wrapped: f.validators[l-1]}
	}

	return f
}

type dateMissingValidator struct {
	label string
}

func (v dateMissingValidator) Validate(value date.Date) Error {
	e := dateMissingError{Label: v.label}

	if value.DayString() == "" {
		e.MissingDay = true
	}
	if value.MonthString() == "" {
		e.MissingMonth = true
	}
	if value.YearString() == "" {
		e.MissingYear = true
	}

	if e.MissingDay || e.MissingMonth || e.MissingYear {
		if e.MissingDay && e.MissingMonth && e.MissingYear {
			return newEmptyError(v.label)
		}

		return e
	}

	return nil
}

func (f *Date) NotEmpty() *Date {
	f.validators = append(f.validators, dateMissingValidator{label: f.Field.Label})

	return f
}

type dateRealValidator struct {
	label string
}

func (v dateRealValidator) Validate(value date.Date) Error {
	if !value.Valid() {
		return newDateMustBeRealError(v.label)
	}

	return nil
}

func (f *Date) Real() *Date {
	f.validators = append(f.validators, dateRealValidator{label: f.Field.Label})

	return f
}

type datePastValidator struct {
	label string
}

func (v datePastValidator) Validate(value date.Date) Error {
	if value.After(date.Today()) {
		return newDateMustBePastError(v.label)
	}

	return nil
}

func (f *Date) Past() *Date {
	f.validators = append(f.validators, datePastValidator{label: f.Field.Label})

	return f
}

type dateBeforeYearsValidator struct {
	label string
	years int
}

func (v dateBeforeYearsValidator) Validate(value date.Date) Error {
	if value.After(date.Today().AddDate(-v.years, 0, 0)) {
		return newDateMustBeBeforeYearsError(v.label, v.years)
	}

	return nil
}

func (f *Date) BeforeYears(years int) *Date {
	f.validators = append(f.validators, dateBeforeYearsValidator{label: f.Field.Label, years: years})

	return f
}

func (f *Date) ErrorPart(part string) bool {
	if err, ok := f.Error.(dateMissingError); ok {
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
