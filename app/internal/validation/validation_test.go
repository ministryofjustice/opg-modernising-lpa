package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	// default state
	var list List
	assert.False(t, list.Any())
	assert.False(t, list.Has("firstName"))
	assert.Equal(t, "", list.Get("firstName"))

	list.Add("firstName", "tooShort")
	assert.True(t, list.Any())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "tooShort", list.Get("firstName"))

	// does not overwrite
	list.Add("firstName", "tooLong")
	assert.True(t, list.Any())
	assert.True(t, list.Has("firstName"))
	assert.Equal(t, "tooShort", list.Get("firstName"))

	var errors []string
	for _, field := range list {
		errors = append(errors, field.Error)
	}
	assert.Equal(t, []string{"tooShort"}, errors)
}
