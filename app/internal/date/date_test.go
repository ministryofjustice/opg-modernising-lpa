package date

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	expected := time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)

	testCases := map[string]struct {
		year, month, day string
		date             Date
	}{
		"unpadded": {
			year:  "2000",
			month: "1",
			day:   "2",
			date:  Date{day: "2", month: "1", year: "2000", t: expected},
		},
		"padded": {
			year:  "2000",
			month: "01",
			day:   "02",
			date:  Date{day: "02", month: "01", year: "2000", t: expected},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			date := New(tc.year, tc.month, tc.day)

			assert.Equal(t, tc.date, date)
			assert.Equal(t, tc.year, date.Year())
			assert.Equal(t, tc.month, date.Month())
			assert.Equal(t, tc.day, date.Day())
			assert.True(t, date.Valid())
			assert.False(t, date.IsZero())
			assert.Equal(t, "2000-1-2", date.String())
		})
	}
}

func TestNewWhenZero(t *testing.T) {
	assert.True(t, New("", "", "").IsZero())
}

func TestNewWhenError(t *testing.T) {
	assert.False(t, New("2000", "100", "2").Valid())
}

func TestToday(t *testing.T) {
	assert.Equal(t, Today().String(), time.Now().Format(dateFormat))
}

func TestBefore(t *testing.T) {
	a := New("1999", "12", "31")
	b := New("2000", "1", "1")

	assert.False(t, a.Before(a))
	assert.True(t, a.Before(b))
	assert.False(t, b.Before(a))
}

func TestAfter(t *testing.T) {
	a := New("1999", "12", "31")
	b := New("2000", "1", "1")

	assert.False(t, a.After(a))
	assert.False(t, a.After(b))
	assert.True(t, b.After(a))
}

func TestAddDate(t *testing.T) {
	a := New("1999", "12", "31")
	b := New("2000", "1", "1")
	c := New("2000", "1", "31")
	d := New("2000", "12", "31")
	e := New("2001", "2", "1")

	assert.Equal(t, b, a.AddDate(0, 0, 1))
	assert.Equal(t, c, a.AddDate(0, 1, 0))
	assert.Equal(t, d, a.AddDate(1, 0, 0))
	assert.Equal(t, e, a.AddDate(1, 1, 1))
}

func TestMarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		date Date
		json string
	}{
		"unpadded": {
			date: New("2020", "5", "21"),
			json: `"2020-5-21"`,
		},
		"padded": {
			date: New("2020", "05", "01"),
			json: `"2020-5-1"`,
		},
		"zero value": {
			json: `""`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			jsonResult, _ := json.Marshal(tc.date)
			assert.Equal(t, []byte(tc.json), jsonResult)
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		json string
		date Date
		err  error
	}{
		"unpadded": {
			json: `"2020-5-21"`,
			date: New("2020", "5", "21"),
		},
		"padded": {
			json: `"2020-05-01"`,
			date: New("2020", "05", "01"),
		},
		"time.RFC3339": {
			json: `"2020-05-01T12:01:02Z"`,
			date: New("2020", "05", "01"),
		},
		"zero value": {
			json: `""`,
		},
		"null": {
			json: `null`,
		},
		"wrong format": {
			json: `"2020-2020-2020-2020"`,
			err:  FormatError("2020-2020-2020-2020"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var dateResult Date
			err := json.Unmarshal([]byte(tc.json), &dateResult)
			assert.Equal(t, tc.date, dateResult)
			assert.Equal(t, tc.err, err)
		})
	}
}
