package identity

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
)

func TestMockClient(t *testing.T) {
	client, err := NewYotiClient("xyz", "", []byte("hey"))
	assert.Nil(t, err)
	assert.True(t, client.IsTest())
	assert.Equal(t, "xyz", client.ScenarioID())

	user, err := client.User("xyz")
	assert.Nil(t, err)
	assert.True(t, user.OK)
	assert.Equal(t, EasyID, user.Provider)
	assert.Equal(t, "Test", user.FirstNames)
	assert.Equal(t, "Person", user.LastName)
	assert.Equal(t, date.New("2000", "1", "2"), user.DateOfBirth)
}
