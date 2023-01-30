package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	// default state
	var list List
	assert.False(t, list.Any())
	assert.True(t, list.None())
	assert.False(t, list.Has("firstName"))
	assert.Equal(t, "", list.Get("firstName"))

	list.Add("firstName", "tooShort")
	assert.True(t, list.Any())
	assert.False(t, list.None())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "tooShort", list.Get("firstName"))

	// does not overwrite
	list.Add("firstName", "tooLong")
	assert.True(t, list.Any())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "tooShort", list.Get("firstName"))

	// ordered
	list.Add("lastName", "tooLong")

	assert.Equal(t, []string{"firstName:tooShort", "lastName:tooLong"}, flatten(list))

	// with is equivalent
	with := With("firstName", "tooShort").
		With("firstName", "tooLong").
		With("lastName", "tooLong")

	assert.Equal(t, list, with)
}

func flatten(l List) []string {
	var s []string
	for _, field := range l {
		s = append(s, fmt.Sprintf("%s:%s", field.Name, field.Error))
	}
	return s
}
