package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCheckYourName(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
					{ID: "other", FirstNames: "Dave", LastName: "Smith"},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			template := newMockTemplate(t)
			template.
				On("Execute", w, &checkYourNameData{
					App:      tc.appData,
					Form:     &checkYourNameForm{},
					Lpa:      tc.lpa,
					Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			err := CheckYourName(template.Execute, donorStore, attorneyStore, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetCheckYourNameWhenAttorneyDoesNotExist(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{
					{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			err := CheckYourName(nil, donorStore, nil, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.Start, resp.Header.Get("Location"))
		})
	}
}

func TestGetCheckYourNameOnDonorStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := CheckYourName(template.Execute, donorStore, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourNameOnAttorneyStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Attorneys: actor.Attorneys{{ID: "attorney-id"}},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	err := CheckYourName(template.Execute, donorStore, attorneyStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourNameOnTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		Attorneys: actor.Attorneys{{ID: "attorney-id"}},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourNameData{
			App:      testAppData,
			Form:     &checkYourNameForm{},
			Lpa:      lpa,
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	err := CheckYourName(template.Execute, donorStore, attorneyStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourName(t *testing.T) {
	testcases := map[string]struct {
		appData         page.AppData
		lpa             *page.Lpa
		attorney        *actor.AttorneyProvidedDetails
		updatedAttorney *actor.AttorneyProvidedDetails
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "yes",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "yes",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"is-name-correct": {"yes"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(tc.attorney, nil)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			err := CheckYourName(nil, donorStore, attorneyStore, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.DateOfBirth, resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourNameWithCorrectedName(t *testing.T) {
	testcases := map[string]struct {
		appData         page.AppData
		lpa             *page.Lpa
		attorney        *actor.AttorneyProvidedDetails
		updatedAttorney *actor.AttorneyProvidedDetails
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Donor:     actor.Donor{Email: "a@example.com"},
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Donor:                actor.Donor{Email: "a@example.com"},
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"is-name-correct": {"no"},
				"corrected-name":  {"Bobby Smith"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(tc.attorney, nil)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", notify.AttorneyNameChangeEmail).
				Return("abc-123")
			notifyClient.
				On("Email", r.Context(), notify.Email{
					EmailAddress:    "a@example.com",
					TemplateID:      "abc-123",
					Personalisation: map[string]string{"declaredName": "Bobby Smith"},
				}).
				Return("", nil)

			err := CheckYourName(nil, donorStore, attorneyStore, notifyClient)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.DateOfBirth, resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourNameWithUnchangedCorrectedName(t *testing.T) {
	testcases := map[string]struct {
		appData         page.AppData
		lpa             *page.Lpa
		attorney        *actor.AttorneyProvidedDetails
		updatedAttorney *actor.AttorneyProvidedDetails
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Donor:     actor.Donor{Email: "a@example.com"},
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				Donor:                actor.Donor{Email: "a@example.com"},
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			},
			attorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				IsNameCorrect: "no",
				CorrectedName: "Bobby Smith",
				Tasks:         actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"is-name-correct": {"no"},
				"corrected-name":  {"Bobby Smith"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			w := httptest.NewRecorder()

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(tc.lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Get", r.Context()).
				Return(tc.attorney, nil)
			attorneyStore.
				On("Put", r.Context(), tc.updatedAttorney).
				Return(nil)

			err := CheckYourName(nil, donorStore, attorneyStore, nil)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.DateOfBirth, resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourNameWithCorrectedNameWhenStoreError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"Bobby Smith"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.AttorneyNameChangeEmail).
		Return("abc-123")
	notifyClient.
		On("Email", r.Context(), notify.Email{
			EmailAddress:    "a@example.com",
			TemplateID:      "abc-123",
			Personalisation: map[string]string{"declaredName": "Bobby Smith"},
		}).
		Return("", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{
			Donor:     actor.Donor{Email: "a@example.com"},
			Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
		}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := CheckYourName(nil, donorStore, attorneyStore, notifyClient)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourNameOnValidationError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {""},
	}

	lpa := &page.Lpa{
		Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourNameData{
			App:      testAppData,
			Form:     &checkYourNameForm{IsNameCorrect: "no"},
			Lpa:      lpa,
			Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
			Errors:   validation.With("corrected-name", validation.EnterError{Label: "yourFullName"}),
		}).
		Return(nil)

	err := CheckYourName(template.Execute, donorStore, attorneyStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadCheckYourNameForm(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"a name"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	assert.Equal(t, &checkYourNameForm{
		IsNameCorrect: "no",
		CorrectedName: "a name",
	},
		readCheckYourNameForm(r),
	)
}

func TestValidateCheckYourNameForm(t *testing.T) {
	testCases := map[string]struct {
		form   checkYourNameForm
		errors validation.List
	}{
		"valid - name correct": {
			form: checkYourNameForm{
				IsNameCorrect: "yes",
			},
			errors: validation.List{},
		},
		"valid - corrected name": {
			form: checkYourNameForm{
				IsNameCorrect: "no",
				CorrectedName: "a name",
			},
			errors: validation.List{},
		},
		"incorrect name missing corrected name": {
			form: checkYourNameForm{
				IsNameCorrect: "no",
			},
			errors: validation.With("corrected-name", validation.EnterError{Label: "yourFullName"}),
		},
		"missing values": {
			form:   checkYourNameForm{},
			errors: validation.With("is-name-correct", validation.CustomError{Label: "confirmIfTheNameIsCorrect"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
