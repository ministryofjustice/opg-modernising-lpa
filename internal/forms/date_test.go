package forms

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/stretchr/testify/assert"
)

func TestDate_Set(t *testing.T) {
	e := NewDate("a", "A")

	e.Set(date.New("2020", "01", "02"))
	assert.Equal(t, "", e.Input)
	assert.Equal(t, "2020", e.InputYear)
	assert.Equal(t, "01", e.InputMonth)
	assert.Equal(t, "02", e.InputDay)

	e.Set(date.New("2020", "1", "2"))
	assert.Equal(t, "", e.Input)
	assert.Equal(t, "2020", e.InputYear)
	assert.Equal(t, "1", e.InputMonth)
	assert.Equal(t, "2", e.InputDay)
}

func TestDate_ParsePostForm(t *testing.T) {
	type formType struct {
		Form
		D *Date
	}

	t.Run("WithError", func(t *testing.T) {
		aForm := formType{
			D: NewDate("a", "A").
				NotEmpty().
				WithError(newEmptyError("nope")),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  2020  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.True(t, aForm.ParsePostForm(req, aForm.D))
			assert.Nil(t, aForm.D.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, newEmptyError("nope"), aForm.D.Error)
		})
	})

	t.Run("NotEmpty", func(t *testing.T) {
		aForm := formType{
			D: NewDate("b", "B").NotEmpty(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  2020  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.True(t, aForm.ParsePostForm(req, aForm.D))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, "2020", aForm.D.InputYear)
			assert.Equal(t, "1", aForm.D.InputMonth)
			assert.Equal(t, "2", aForm.D.InputDay)
			assert.Equal(t, date.New("2020", "1", "2"), aForm.D.Value)
			assert.Nil(t, aForm.D.Error)
		})

		t.Run("missing all", func(t *testing.T) {
			req := makeRequest(url.Values{})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New("", "", ""), aForm.D.Value)
			assert.Equal(t, newEmptyError("B"), aForm.D.Error)
		})

		t.Run("missing year", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New("", "1", "2"), aForm.D.Value)
			assert.Equal(t, dateMissingError{Label: "B", MissingYear: true}, aForm.D.Error)
		})

		t.Run("missing month and day", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year": {"  2020  "},
			})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New("2020", "", ""), aForm.D.Value)
			assert.Equal(t, dateMissingError{Label: "B", MissingMonth: true, MissingDay: true}, aForm.D.Error)
		})
	})

	t.Run("Real", func(t *testing.T) {
		aForm := formType{
			D: NewDate("b", "B").Real(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  2020  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.True(t, aForm.ParsePostForm(req, aForm.D))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, "2020", aForm.D.InputYear)
			assert.Equal(t, "1", aForm.D.InputMonth)
			assert.Equal(t, "2", aForm.D.InputDay)
			assert.Equal(t, date.New("2020", "1", "2"), aForm.D.Value)
			assert.Nil(t, aForm.D.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  hey  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New("hey", "1", "2"), aForm.D.Value)
			assert.Equal(t, newDateMustBeRealError("B"), aForm.D.Error)
		})
	})

	t.Run("Past", func(t *testing.T) {
		aForm := formType{
			D: NewDate("b", "B").Past(),
		}

		t.Run("valid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  2020  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.True(t, aForm.ParsePostForm(req, aForm.D))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, "2020", aForm.D.InputYear)
			assert.Equal(t, "1", aForm.D.InputMonth)
			assert.Equal(t, "2", aForm.D.InputDay)
			assert.Equal(t, date.New("2020", "1", "2"), aForm.D.Value)
			assert.Nil(t, aForm.D.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {"  2999  "},
				aForm.D.Name + "-month": {"  1  "},
				aForm.D.Name + "-day":   {"  2  "},
			})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New("2999", "1", "2"), aForm.D.Value)
			assert.Equal(t, newDateMustBePastError("B"), aForm.D.Error)
		})
	})

	t.Run("BeforeYears", func(t *testing.T) {
		aForm := formType{
			D: NewDate("b", "B").BeforeYears(1),
		}

		t.Run("valid", func(t *testing.T) {
			lastYear := time.Now().AddDate(-1, 0, -1)
			y, m, d := strconv.Itoa(lastYear.Year()), strconv.Itoa(int(lastYear.Month())), strconv.Itoa(lastYear.Day())

			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {y},
				aForm.D.Name + "-month": {m},
				aForm.D.Name + "-day":   {d},
			})

			assert.True(t, aForm.ParsePostForm(req, aForm.D))
			assert.Empty(t, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, y, aForm.D.InputYear)
			assert.Equal(t, m, aForm.D.InputMonth)
			assert.Equal(t, d, aForm.D.InputDay)
			assert.Equal(t, date.New(y, m, d), aForm.D.Value)
			assert.Nil(t, aForm.D.Error)
		})

		t.Run("invalid", func(t *testing.T) {
			lastYear := time.Now().AddDate(-1, 0, 1)
			y, m, d := strconv.Itoa(lastYear.Year()), strconv.Itoa(int(lastYear.Month())), strconv.Itoa(lastYear.Day())

			req := makeRequest(url.Values{
				aForm.D.Name + "-year":  {y},
				aForm.D.Name + "-month": {m},
				aForm.D.Name + "-day":   {d},
			})

			assert.False(t, aForm.ParsePostForm(req, aForm.D))
			assert.Equal(t, []Field{aForm.D.Field}, aForm.Errors)
			assert.Equal(t, "", aForm.D.Input)
			assert.Equal(t, date.New(y, m, d), aForm.D.Value)
			assert.Equal(t, newDateMustBeBeforeYearsError("B", 1), aForm.D.Error)
		})
	})
}
