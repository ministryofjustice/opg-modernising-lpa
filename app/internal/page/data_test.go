package page

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadDate(t *testing.T) {
	date := readDate(time.Date(2020, time.March, 12, 0, 0, 0, 0, time.Local))

	assert.Equal(t, Date{Day: "12", Month: "3", Year: "2020"}, date)
}
