package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	l := newMockLocalizer(t)
	l.On("T", "a").Return("A")
	l.On("T", "c").Return("C")
	l.On("Format", "errorStringTooLong", map[string]any{"Label": "A", "Length": 4}).Return("a-tooLong")
	l.On("Format", "errorStringTooLong", map[string]any{"Label": "C", "Length": 3}).Return("c-tooLong")

	// default state
	var list List
	assert.False(t, list.Any())
	assert.True(t, list.None())
	assert.False(t, list.Has("firstName"))
	assert.Equal(t, "", list.Format(l, "firstName"))

	list.Add("firstName", StringTooLongError{Label: "a", Length: 4})
	assert.True(t, list.Any())
	assert.False(t, list.None())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "a-tooLong", list.Format(l, "firstName"))

	// does not overwrite
	list.Add("firstName", StringLengthError{Label: "b"})
	assert.True(t, list.Any())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "a-tooLong", list.Format(l, "firstName"))

	// ordered
	list.Add("lastName", StringTooLongError{Label: "c", Length: 3})

	assert.Equal(t, []string{"firstName:a-tooLong", "lastName:c-tooLong"}, flatten(l, list))

	// with is equivalent
	with := With("firstName", StringTooLongError{Label: "a", Length: 4}).
		With("firstName", StringLengthError{Label: "b"}).
		With("lastName", StringTooLongError{Label: "c", Length: 3})

	assert.Equal(t, list, with)
}

func TestValidationForDates(t *testing.T) {
	var list List
	assert.True(t, list.None())

	list.Add("day", DateMissingError{Label: "a", MissingDay: true})
	list.Add("month", DateMissingError{Label: "b", MissingMonth: true})
	list.Add("year", DateMissingError{Label: "c", MissingYear: true})
	list.Add("enter", EnterError{Label: "d"})

	assert.True(t, list.HasForDate("day", "day"))
	assert.False(t, list.HasForDate("day", "month"))
	assert.False(t, list.HasForDate("day", "year"))

	assert.False(t, list.HasForDate("month", "day"))
	assert.True(t, list.HasForDate("month", "month"))
	assert.False(t, list.HasForDate("month", "year"))

	assert.False(t, list.HasForDate("year", "day"))
	assert.False(t, list.HasForDate("year", "month"))
	assert.True(t, list.HasForDate("year", "year"))

	assert.True(t, list.HasForDate("enter", "day"))
	assert.True(t, list.HasForDate("enter", "month"))
	assert.True(t, list.HasForDate("enter", "year"))
}

func flatten(l Localizer, list List) []string {
	var s []string
	for _, field := range list {
		s = append(s, fmt.Sprintf("%s:%s", field.Name, field.Error.Format(l)))
	}
	return s
}
