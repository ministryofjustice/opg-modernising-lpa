package donordata

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
)

func TestDonorFullName(t *testing.T) {
	d := Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}
	whiteSpace := Donor{FirstNames: " ", LastName: "  "}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
	assert.Equal(t, "", whiteSpace.FullName())
}

func TestDonorIsUnder18(t *testing.T) {
	assert.True(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, 1)}.IsUnder18())
	assert.False(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, 0)}.IsUnder18())
	assert.False(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, -1)}.IsUnder18())
}

func TestDonorIs18On(t *testing.T) {
	assert.False(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, 1)}.Is18On(time.Now()))
	assert.True(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, 0)}.Is18On(time.Now()))
	assert.True(t, Donor{DateOfBirth: date.Today().AddDate(-18, 0, -1)}.Is18On(time.Now()))
}
