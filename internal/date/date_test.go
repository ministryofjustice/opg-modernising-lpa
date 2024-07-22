package date

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	expected := time.Date(2000, time.March, 4, 0, 0, 0, 0, time.UTC)

	testCases := map[string]struct {
		year, month, day string
		date             Date
	}{
		"unpadded": {
			year:  "2000",
			month: "3",
			day:   "4",
			date:  Date{day: "4", month: "3", year: "2000", t: expected},
		},
		"padded": {
			year:  "2000",
			month: "03",
			day:   "04",
			date:  Date{day: "04", month: "03", year: "2000", t: expected},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			date := New(tc.year, tc.month, tc.day)

			assert.Equal(t, tc.date, date)
			assert.Equal(t, tc.year, date.YearString())
			assert.Equal(t, 2000, date.Year())
			assert.Equal(t, tc.month, date.MonthString())
			assert.Equal(t, time.March, date.Month())
			assert.Equal(t, tc.day, date.DayString())
			assert.Equal(t, 4, date.Day())
			assert.True(t, date.Valid())
			assert.False(t, date.IsZero())
			assert.Equal(t, "2000-3-4", date.String())
			assert.Equal(t, "4 March 2000", date.Format("2 January 2006"))
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
	assert.Equal(t, Today().String(), time.Now().Format(unpaddedDate))
}

func TestFromTime(t *testing.T) {
	assert.Equal(t, New("2000", "01", "02"), FromTime(time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, Date{}, FromTime(time.Time{}))
}

func TestEquals(t *testing.T) {
	a := New("1999", "12", "31")
	b := New("2000", "1", "1")
	c := New("2000", "01", "01")

	assert.True(t, a.Equals(a))
	assert.False(t, a.Equals(b))
	assert.False(t, b.Equals(a))
	assert.True(t, b.Equals(c))
	assert.True(t, c.Equals(b))
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
	b := New("2000", "01", "01")
	c := New("2000", "01", "31")
	d := New("2000", "12", "31")
	e := New("2001", "02", "01")

	assert.Equal(t, b, a.AddDate(0, 0, 1))
	assert.Equal(t, c, a.AddDate(0, 1, 0))
	assert.Equal(t, d, a.AddDate(1, 0, 0))
	assert.Equal(t, e, a.AddDate(1, 1, 1))
}

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		date Date
		json string
		av   types.AttributeValue
	}{
		"unpadded": {
			date: New("2020", "5", "21"),
			json: `"2020-05-21"`,
			av:   &types.AttributeValueMemberS{Value: "2020-5-21"},
		},
		"padded": {
			date: New("2020", "05", "01"),
			json: `"2020-05-01"`,
			av:   &types.AttributeValueMemberS{Value: "2020-5-1"},
		},
		"zero value": {
			json: `""`,
			av:   &types.AttributeValueMemberS{Value: ""},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			jsonResult, err := json.Marshal(tc.date)
			assert.Nil(t, err)
			assert.Equal(t, []byte(tc.json), jsonResult)

			avResult, err := attributevalue.Marshal(tc.date)
			assert.Nil(t, err)
			assert.Equal(t, tc.av, avResult)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	expectedTime := time.Date(2020, time.May, 1, 12, 1, 2, 0, time.UTC)
	jsonTime, _ := json.Marshal(expectedTime)
	avTime, _ := attributevalue.Marshal(expectedTime)

	testCases := map[string]struct {
		json      string
		av        types.AttributeValue
		date      Date
		err       error
		typeError bool
	}{
		"unpadded": {
			json: `"2020-5-21"`,
			av:   &types.AttributeValueMemberS{Value: "2020-5-21"},
			date: New("2020", "5", "21"),
		},
		"padded": {
			json: `"2020-05-01"`,
			av:   &types.AttributeValueMemberS{Value: "2020-05-01"},
			date: New("2020", "05", "01"),
		},
		"time.Time": {
			json: string(jsonTime),
			av:   avTime,
			date: New("2020", "05", "01"),
		},
		"zero value": {
			json: `""`,
			av:   &types.AttributeValueMemberS{Value: ""},
		},
		"null": {
			json: `null`,
			av:   &types.AttributeValueMemberNULL{Value: true},
		},
		"wrong format": {
			json: `"2020-2020-2020-2020"`,
			av:   &types.AttributeValueMemberS{Value: "2020-2020-2020-2020"},
			err:  FormatError("2020-2020-2020-2020"),
		},
		"wrong type": {
			json:      `123`,
			av:        &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
			typeError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name+" json", func(t *testing.T) {
			var date Date
			err := json.Unmarshal([]byte(tc.json), &date)
			assert.Equal(t, tc.date, date)
			if tc.typeError {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}
		})

		t.Run(name+" attributevalue", func(t *testing.T) {
			var date Date
			err := attributevalue.Unmarshal(tc.av, &date)
			assert.Equal(t, tc.date, date)
			if tc.typeError {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.err, err)
			}
		})
	}
}

func TestTime(t *testing.T) {
	date := New("2000", "1", "2")
	assert.Equal(t, time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC), date.Time())
}

func TestHash(t *testing.T) {
	date := New("2000", "1", "2")
	hash, err := date.Hash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x386e9500), hash)

	date = New("2001", "1", "2")
	hash, err = date.Hash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x3a511a00), hash)
}
