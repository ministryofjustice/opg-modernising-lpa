package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type ErrorChecker interface {
	CheckError(label string, value error) FormattableError
}

func (l *List) Error(name, label string, value error, checks ...ErrorChecker) {
	for _, check := range checks {
		if err := check.CheckError(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type StringChecker interface {
	CheckString(label, value string) FormattableError
}

func (l *List) String(name, label, value string, checks ...StringChecker) {
	for _, check := range checks {
		if err := check.CheckString(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type DateChecker interface {
	CheckDate(string, date.Date) FormattableError
}

func (l *List) Date(name, label string, value date.Date, checks ...DateChecker) {
	for _, check := range checks {
		if err := check.CheckDate(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type AddressChecker interface {
	CheckAddress(string, *place.Address) FormattableError
}

func (l *List) Address(name, label string, value *place.Address, checks ...AddressChecker) {
	for _, check := range checks {
		if err := check.CheckAddress(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type BoolChecker interface {
	CheckBool(string, bool) FormattableError
}

func (l *List) Bool(name, label string, value bool, checks ...BoolChecker) {
	for _, check := range checks {
		if err := check.CheckBool(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type OptionsChecker interface {
	CheckOptions(string, []string) FormattableError
}

func (l *List) Options(name, label string, value []string, checks ...OptionsChecker) {
	for _, check := range checks {
		if err := check.CheckOptions(label, value); err != nil {
			l.Add(name, err)
			return
		}
	}
}

type SelectedCheck struct {
	useCustomError bool
}

func (c SelectedCheck) CustomError() SelectedCheck {
	c.useCustomError = true
	return c
}

func (c SelectedCheck) CheckBool(label string, value bool) FormattableError {
	if !value {
		if c.useCustomError {
			return CustomError{Label: label}
		} else {
			return SelectError{Label: label}
		}
	}

	return nil
}

func (c SelectedCheck) CheckOptions(label string, value []string) FormattableError {
	if len(value) == 0 {
		if c.useCustomError {
			return CustomError{Label: label}
		} else {
			return SelectError{Label: label}
		}
	}

	return nil
}

func (c SelectedCheck) CheckAddress(label string, value *place.Address) FormattableError {
	if value == nil {
		if c.useCustomError {
			return CustomError{Label: label}
		} else {
			return SelectError{Label: label}
		}
	}

	return nil
}

func (c SelectedCheck) CheckError(label string, err error) FormattableError {
	if err != nil {
		return SelectError{Label: label}
	}

	return nil
}

func Selected() SelectedCheck {
	return SelectedCheck{}
}

type SelectCheck struct {
	in             []string
	useCustomError bool
}

func (c SelectCheck) CustomError() SelectCheck {
	c.useCustomError = true
	return c
}

func (c SelectCheck) CheckString(label string, value string) FormattableError {
	if !slices.Contains(c.in, value) {
		if c.useCustomError {
			return CustomError{Label: label}
		} else {
			return SelectError{Label: label}
		}
	}

	return nil
}

func (c SelectCheck) CheckOptions(label string, value []string) FormattableError {
	for _, v := range value {
		if !slices.Contains(c.in, v) {
			if c.useCustomError {
				return CustomError{Label: label}
			} else {
				return SelectError{Label: label}
			}
		}
	}

	return nil
}

func Select(in ...string) SelectCheck {
	return SelectCheck{in: in}
}

type EmptyCheck struct{}

func (c EmptyCheck) CheckString(label, value string) FormattableError {
	if value == "" {
		return EnterError{Label: label}
	}

	return nil
}

func Empty() EmptyCheck {
	return EmptyCheck{}
}

type StringTooLongCheck struct {
	length int
}

func (c StringTooLongCheck) CheckString(label, value string) FormattableError {
	if len(value) > c.length {
		return StringTooLongError{Label: label, Length: c.length}
	}

	return nil
}

func StringTooLong(length int) StringTooLongCheck {
	return StringTooLongCheck{length: length}
}

type StringLengthCheck struct {
	length int
}

func (c StringLengthCheck) CheckString(label, value string) FormattableError {
	if len(value) != c.length {
		return StringLengthError{Label: label, Length: c.length}
	}

	return nil
}

func StringLength(length int) StringLengthCheck {
	return StringLengthCheck{length: length}
}

var (
	MobileRegex      = regexp.MustCompile(`^(?:07|\+?447)\d{9}$`)
	NonUKMobileRegex = regexp.MustCompile(`^\+\d{4,15}$`)
)

type MobileCheck struct {
	nonUK      bool
	errorLabel string
}

func (c MobileCheck) ErrorLabel(label string) MobileCheck {
	c.errorLabel = label
	return c
}

func (c MobileCheck) CheckString(label, value string) FormattableError {
	re := MobileRegex
	if c.nonUK {
		re = NonUKMobileRegex
	}

	if value != "" && !re.MatchString(strings.ReplaceAll(value, " ", "")) {
		if c.errorLabel != "" {
			return CustomError{Label: c.errorLabel}
		} else {
			return MobileError{Label: label}
		}
	}

	return nil
}

func Mobile() MobileCheck {
	return MobileCheck{}
}

func NonUKMobile() MobileCheck {
	return MobileCheck{nonUK: true}
}

type EmailCheck struct{}

func (c EmailCheck) CheckString(label, value string) FormattableError {
	if value != "" {
		if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", value)); err != nil {
			return EmailError{Label: label}
		}
	}

	return nil
}

func Email() EmailCheck {
	return EmailCheck{}
}

type DateMissingCheck struct{}

func (c DateMissingCheck) CheckDate(label string, date date.Date) FormattableError {
	e := DateMissingError{Label: label}

	if date.DayString() == "" {
		e.MissingDay = true
	}
	if date.MonthString() == "" {
		e.MissingMonth = true
	}
	if date.YearString() == "" {
		e.MissingYear = true
	}

	if e.MissingDay || e.MissingMonth || e.MissingYear {
		if e.MissingDay && e.MissingMonth && e.MissingYear {
			return EnterError{Label: label}
		}

		return e
	}

	return nil
}

func DateMissing() DateMissingCheck {
	return DateMissingCheck{}
}

type DateMustBeRealCheck struct{}

func (c DateMustBeRealCheck) CheckDate(label string, value date.Date) FormattableError {
	if !value.Valid() {
		return DateMustBeRealError{Label: label}
	}

	return nil
}

func DateMustBeReal() DateMustBeRealCheck {
	return DateMustBeRealCheck{}
}

type DateMustBePastCheck struct{}

func (c DateMustBePastCheck) CheckDate(label string, value date.Date) FormattableError {
	if value.After(date.Today()) {
		return DateMustBePastError{Label: label}
	}

	return nil
}

func DateMustBePast() DateMustBePastCheck {
	return DateMustBePastCheck{}
}
