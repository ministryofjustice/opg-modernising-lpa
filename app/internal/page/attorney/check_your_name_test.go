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
)

func TestGetCheckYourName(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lpa := &page.Lpa{
		Attorneys: actor.Attorneys{
			{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
			{ID: "other", FirstNames: "Dave", LastName: "Smith"},
		},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourNameData{
			App:      testAppData,
			Form:     &checkYourNameForm{},
			Lpa:      lpa,
			Attorney: actor.Attorney{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"},
		}).
		Return(nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	err := CheckYourName(template.Execute, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourNameWhenAttorneyDoesNotExist(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	err := CheckYourName(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start, resp.Header.Get("Location"))
}

func TestGetCheckYourNameOnStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := CheckYourName(template.Execute, lpaStore, nil)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	err := CheckYourName(template.Execute, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourName(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	err := CheckYourName(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.DateOfBirth, resp.Header.Get("Location"))
}

func TestPostCheckYourNameWithCorrectedName(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"Bobby Smith"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor:     actor.Donor{Email: "a@example.com"},
			Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Donor:                   actor.Donor{Email: "a@example.com"},
			Attorneys:               actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			AttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id", DeclaredFullName: "Bobby Smith"}},
		}).
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

	err := CheckYourName(nil, lpaStore, notifyClient)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.DateOfBirth, resp.Header.Get("Location"))
}

func TestPostCheckYourNameWithCorrectedNameWhenStoreError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"Bobby Smith"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Attorneys:               actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
			AttorneyProvidedDetails: actor.Attorneys{{ID: "attorney-id", DeclaredFullName: "Bobby Smith"}},
		}).
		Return(expectedError)

	err := CheckYourName(nil, lpaStore, nil)(testAppData, w, r)
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

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

	err := CheckYourName(template.Execute, lpaStore, nil)(testAppData, w, r)
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
			errors: validation.With("is-name-correct", validation.SelectError{Label: "yesIfTheNameIsCorrect"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
