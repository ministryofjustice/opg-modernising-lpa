package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaType(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App: testAppData,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App:  testAppData,
			Type: page.LpaTypePropertyFinance,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &page.Lpa{Type: page.LpaTypePropertyFinance})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App: testAppData,
		}).
		Return(expectedError)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLpaType(t *testing.T) {
	form := url.Values{
		"lpa-type": {page.LpaTypePropertyFinance},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
			Type:  page.LpaTypePropertyFinance,
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	err := LpaType(nil, donorStore)(testAppData, w, r, &page.Lpa{
		Donor: actor.Donor{
			FirstNames:  "Jane",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "1", "2"),
			Address:     place.Address{Postcode: "ABC123"},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostLpaTypeWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"lpa-type": {page.LpaTypePropertyFinance},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := LpaType(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostLpaTypeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App:    testAppData,
			Errors: validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}),
		}).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadLpaTypeForm(t *testing.T) {
	form := url.Values{
		"lpa-type": {page.LpaTypePropertyFinance},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readLpaTypeForm(r)

	assert.Equal(t, page.LpaTypePropertyFinance, result.LpaType)
}

func TestLpaTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lpaTypeForm
		errors validation.List
	}{
		"pfa": {
			form: &lpaTypeForm{
				LpaType: page.LpaTypePropertyFinance,
			},
		},
		"hw": {
			form: &lpaTypeForm{
				LpaType: "hw",
			},
		},
		"missing": {
			form:   &lpaTypeForm{},
			errors: validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}),
		},
		"invalid": {
			form: &lpaTypeForm{
				LpaType: "what",
			},
			errors: validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
