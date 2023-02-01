package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	expected := time.Date(2020, time.March, 12, 0, 0, 0, 0, time.UTC)
	date := Read(expected)

	assert.Equal(t, Date{Day: "12", Month: "3", Year: "2020", T: expected}, date)
}

func TestFromParts(t *testing.T) {
	expected := time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)
	date := FromParts("2000", "1", "2")

	assert.Equal(t, Date{Day: "2", Month: "1", Year: "2000", T: expected}, date)
}

func TestFromPartsWhenError(t *testing.T) {
	date := FromParts("2000", "100", "2")

	assert.NotNil(t, date.Err)
}
