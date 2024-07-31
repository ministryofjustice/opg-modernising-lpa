package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourEmail(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourEmailData{
			App:  testAppData,
			Form: &yourEmailForm{},
		}).
		Return(nil)

	err := YourEmail(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourEmailWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourEmail(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourEmail(t *testing.T) {
	form := url.Values{
		"email": {"john@example.com"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames: "John",
				Email:      "john@example.com",
			},
		}).
		Return(nil)

	err := YourEmail(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			FirstNames: "John",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CanYouSignYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourEmailWhenValidationError(t *testing.T) {
	form := url.Values{
		"email": {"john"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *yourEmailData) bool {
			return assert.Equal(t, validation.With("email", validation.EmailError{Label: "email"}), data.Errors)
		})).
		Return(nil)

	err := YourEmail(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourEmailWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"email": {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := YourEmail(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestReadYourEmailForm(t *testing.T) {
	form := url.Values{"email": {"john@example.com"}}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourEmailForm(r)

	assert.Equal(t, "john@example.com", result.Email)
}

func TestYourEmailFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourEmailForm
		errors validation.List
	}{
		"valid": {
			form: &yourEmailForm{Email: "john@example.com"},
		},
		"empty": {
			form: &yourEmailForm{},
		},
		"invalid": {
			form:   &yourEmailForm{Email: "john"},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
