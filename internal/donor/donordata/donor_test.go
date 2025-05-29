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

func TestDonorNameHasChanged(t *testing.T) {
	testCases := map[string]*Donor{
		"FirstNames": {FirstNames: "d", LastName: "b", OtherNames: "c"},
		"LastName":   {FirstNames: "a", LastName: "d", OtherNames: "c"},
		"OtherNames": {FirstNames: "a", LastName: "b", OtherNames: "d"},
	}

	donor := Donor{FirstNames: "a", LastName: "b", OtherNames: "c"}

	for name, updatedDonor := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.True(t, donor.NameHasChanged(updatedDonor.FirstNames, updatedDonor.LastName, updatedDonor.OtherNames))
		})
	}

	assert.False(t, donor.NameHasChanged("a", "b", "c"))
}
