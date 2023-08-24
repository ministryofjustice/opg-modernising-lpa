package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPreviousApplicationNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &previousApplicationNumberData{
			App:  testAppData,
			Form: &previousApplicationNumberForm{},
		}).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousApplicationNumberFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &previousApplicationNumberData{
			App: testAppData,
			Form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "ABC",
			},
		}).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{PreviousApplicationNumber: "ABC"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousApplicationNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := PreviousApplicationNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPreviousApplicationNumber(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"ABC"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                "lpa-id",
			UID:               "lpa-uid",
			ApplicationReason: page.AdditionalApplication,
			Donor: actor.Donor{
				FirstNames:  "Jane",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{Postcode: "ABC123"},
			},
			PreviousApplicationNumber: "ABC",
			Tasks:                     page.Tasks{YourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	err := PreviousApplicationNumber(nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID:                "lpa-id",
		UID:               "lpa-uid",
		ApplicationReason: page.AdditionalApplication,
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
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostPreviousApplicationNumberWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"ABC"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := PreviousApplicationNumber(nil, donorStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostPreviousApplicationNumberWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *previousApplicationNumberData) bool {
			return assert.Equal(t, validation.With("previous-application-number", validation.EnterError{Label: "previousApplicationNumber"}), data.Errors)
		})).
		Return(nil)

	err := PreviousApplicationNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadPreviousApplicationNumberForm(t *testing.T) {
	form := url.Values{
		"previous-application-number": {"ABC"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readPreviousApplicationNumberForm(r)

	assert.Equal(t, "ABC", result.PreviousApplicationNumber)
}

func TestPreviousApplicationNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *previousApplicationNumberForm
		errors validation.List
	}{
		"valid": {
			form: &previousApplicationNumberForm{
				PreviousApplicationNumber: "A",
			},
		},
		"empty": {
			form:   &previousApplicationNumberForm{},
			errors: validation.With("previous-application-number", validation.EnterError{Label: "previousApplicationNumber"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
