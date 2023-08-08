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

func TestGetApplicationReason(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &applicationReasonData{
			App:     testAppData,
			Form:    &applicationReasonForm{},
			Options: page.ApplicationReasonValues,
		}).
		Return(nil)

	err := ApplicationReason(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetApplicationReasonFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &applicationReasonData{
			App: testAppData,
			Form: &applicationReasonForm{
				ApplicationReason: page.RemakeOfInvalidApplication,
			},
			Options: page.ApplicationReasonValues,
		}).
		Return(nil)

	err := ApplicationReason(template.Execute, nil)(testAppData, w, r, &page.Lpa{ApplicationReason: page.RemakeOfInvalidApplication})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetApplicationReasonWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ApplicationReason(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostApplicationReason(t *testing.T) {
	testcases := map[page.ApplicationReason]struct {
		redirect string
		tasks    page.Tasks
	}{
		page.NewApplication: {
			redirect: page.Paths.TaskList.Format("lpa-id"),
			tasks:    page.Tasks{YourDetails: actor.TaskCompleted},
		},
		page.RemakeOfInvalidApplication: {
			redirect: page.Paths.PreviousApplicationNumber.Format("lpa-id"),
		},
	}

	for reason, tc := range testcases {
		t.Run(reason.String(), func(t *testing.T) {
			form := url.Values{
				"application-reason": {reason.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID: "lpa-id",
					Donor: actor.Donor{
						FirstNames:  "Jane",
						LastName:    "Smith",
						DateOfBirth: date.New("2000", "1", "2"),
						Address:     place.Address{Postcode: "ABC123"},
					},
					ApplicationReason: reason,
					Tasks:             tc.tasks,
				}).
				Return(nil)

			err := ApplicationReason(nil, donorStore)(testAppData, w, r, &page.Lpa{
				ID: "lpa-id",
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
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostApplicationReasonWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"application-reason": {page.NewApplication.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ApplicationReason(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostApplicationReasonWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *applicationReasonData) bool {
			return assert.Equal(t, validation.With("application-reason", validation.SelectError{Label: "theReasonForMakingTheApplication"}), data.Errors)
		})).
		Return(nil)

	err := ApplicationReason(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadApplicationReasonForm(t *testing.T) {
	form := url.Values{
		"application-reason": {page.NewApplication.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readApplicationReasonForm(r)

	assert.Equal(t, page.NewApplication, result.ApplicationReason)
}

func TestApplicationReasonFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *applicationReasonForm
		errors validation.List
	}{
		"valid": {
			form: &applicationReasonForm{},
		},
		"invalid": {
			form: &applicationReasonForm{
				Error: expectedError,
			},
			errors: validation.With("application-reason", validation.SelectError{Label: "theReasonForMakingTheApplication"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
