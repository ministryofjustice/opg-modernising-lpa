package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	client, err := NewYotiClient("", []byte("hey"))
	assert.Nil(t, err)
	assert.True(t, client.IsTest())

	user, err := client.User("xyz")
	assert.Nil(t, err)
	assert.Equal(t, UserData{FullName: "Test Person"}, user)
}
