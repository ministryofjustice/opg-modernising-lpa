package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

type testEnum string

func (e testEnum) Empty() bool { return len(e) == 0 }
func (e *testEnum) UnmarshalText(data []byte) error {
	*e = testEnum(string(data))
	return nil
}

type testEnumOptions struct {
	Current testEnum
	Next    testEnum
}

var testEnumValues = testEnumOptions{
	Current: "current",
	Next:    "next",
}

func TestNewSelectForm(t *testing.T) {
	f := NewSelectForm(testEnumValues.Current, testEnumValues, "aValue")

	assert.Equal(t, &SelectForm[testEnum, testEnumOptions, *testEnum]{
		Selected:   testEnumValues.Current,
		FieldName:  FieldNames.Select,
		Options:    testEnumValues,
		ErrorLabel: "aValue",
	}, f)
}

func TestNewEmptySelectForm(t *testing.T) {
	f := NewEmptySelectForm[testEnum](testEnumValues, "aValue")

	assert.Equal(t, &SelectForm[testEnum, testEnumOptions, *testEnum]{
		FieldName:  FieldNames.Select,
		Options:    testEnumValues,
		ErrorLabel: "aValue",
	}, f)
}

func TestSelectFormRead(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{FieldNames.Select: {"whatever"}}.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	f := NewEmptySelectForm[testEnum](testEnumValues, "aValue")
	f.Read(r)

	assert.Equal(t, testEnum("whatever"), f.Selected)
}

func TestSelectFormValidateWhenValid(t *testing.T) {
	f := NewSelectForm(testEnumValues.Current, testEnumValues, "aValue")
	assert.Nil(t, f.Validate())
}

func TestSelectFormValidateWhenInvalid(t *testing.T) {
	f := NewEmptySelectForm[testEnum](testEnumValues, "aValue")
	assert.Equal(t, validation.With(FieldNames.Select, validation.SelectError{Label: "aValue"}), f.Validate())
}
