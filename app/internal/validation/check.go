package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"golang.org/x/exp/slices"
)

type StringCheck interface {
	CheckString(label, value string) FormattableError
}

func (l *List) String(name, label, value string, checks ...StringCheck) {
	for _, check := range checks {
		if err := check.CheckString(label, value); err != nil {
			l.AddKey(name, err)
			return
		}
	}
}

type DateCheck interface {
	CheckDate(string, date.Date) FormattableError
}

func (l *List) Date(name, label string, value date.Date, checks ...DateCheck) {
	for _, check := range checks {
		if err := check.CheckDate(label, value); err != nil {
			l.AddKey(name, err)
			return
		}
	}
}

type AddressCheck interface {
	CheckAddress(string, *place.Address) FormattableError
}

func (l *List) Address(name, label string, value *place.Address, checks ...AddressCheck) {
	for _, check := range checks {
		if err := check.CheckAddress(label, value); err != nil {
			l.AddKey(name, err)
			return
		}
	}
}

type BoolCheck interface {
	CheckBool(string, bool) FormattableError
}

func (l *List) Bool(name, label string, value bool, checks ...BoolCheck) {
	for _, check := range checks {
		if err := check.CheckBool(label, value); err != nil {
			l.AddKey(name, err)
			return
		}
	}
}

type OptionsCheck interface {
	CheckOptions(string, []string) FormattableError
}

func (l *List) Options(name, label string, value []string, checks ...OptionsCheck) {
	for _, check := range checks {
		if err := check.CheckOptions(label, value); err != nil {
			l.AddKey(name, err)
			return
		}
	}
}

type SelectedCheck struct{}

func (c SelectedCheck) CheckBool(label string, value bool) FormattableError {
	if !value {
		return SelectedError{Label: label}
	}

	return nil
}

func (c SelectedCheck) CheckOptions(label string, value []string) FormattableError {
	if len(value) == 0 {
		return SelectedError{Label: label}
	}

	return nil
}

func Selected() SelectedCheck {
	return SelectedCheck{}
}

type SelectCheck struct {
	in []string
}

func (c SelectCheck) CheckString(label string, value string) FormattableError {
	if !slices.Contains(c.in, value) {
		return SelectError{Label: label}
	}

	return nil
}

func (c SelectCheck) CheckOptions(label string, value []string) FormattableError {
	for _, v := range value {
		if !slices.Contains(c.in, v) {
			return SelectError{Label: label}
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

var mobileRegex = regexp.MustCompile(`^(?:07|\+?447)\d{9}$`)

type MobileCheck struct{}

func (c MobileCheck) CheckString(label, value string) FormattableError {
	if !mobileRegex.MatchString(value) {
		return MobileError{Label: label}
	}

	return nil
}

func Mobile() MobileCheck {
	return MobileCheck{}
}

type EmailCheck struct{}

func (c EmailCheck) CheckString(label, value string) FormattableError {
	if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", value)); err != nil {
		return EmailError{Label: label}
	}

	return nil
}

func Email() EmailCheck {
	return EmailCheck{}
}

type DateMissingCheck struct{}

func (c DateMissingCheck) CheckDate(label string, date date.Date) FormattableError {
	e := DateMissingError{Label: label}

	if date.Day == "" {
		e.MissingDay = true
	}
	if date.Month == "" {
		e.MissingMonth = true
	}
	if date.Year == "" {
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
	if value.TErr != nil {
		return DateMustBeRealError{Label: label}
	}

	return nil
}

func DateMustBeReal() DateMustBeRealCheck {
	return DateMustBeRealCheck{}
}

type DateMustBePastCheck struct{}

func (c DateMustBePastCheck) CheckDate(label string, value date.Date) FormattableError {
	today := time.Now().UTC().Round(24 * time.Hour)

	if value.T.After(today) {
		return DateMustBePastError{Label: label}
	}

	return nil
}

func DateMustBePast() DateMustBePastCheck {
	return DateMustBePastCheck{}
}

type AddressSelectedCheck struct{}

func (c AddressSelectedCheck) CheckAddress(label string, value *place.Address) FormattableError {
	if value == nil {
		return AddressSelectedError{Label: label}
	}

	return nil
}

func AddressSelected() AddressSelectedCheck {
	return AddressSelectedCheck{}
}
