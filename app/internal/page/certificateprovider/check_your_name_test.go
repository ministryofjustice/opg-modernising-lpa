package certificateprovider

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

func TestGetEnterYourName(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	data := checkYourNameData{
		App:  testAppData,
		Form: &checkYourNameForm{},
		Lpa:  lpa,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	err := CheckYourName(template.Execute, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterYourNameOnDonorStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := CheckYourName(template.Execute, donorStore, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterYourNameOnCertificateProviderStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := CheckYourName(template.Execute, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterYourNameOnTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	data := checkYourNameData{
		App:  testAppData,
		Form: &checkYourNameForm{},
		Lpa:  lpa,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	err := CheckYourName(template.Execute, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterYourName(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{FirstNames: "Bob", LastName: "Smith"}).
		Return(nil)

	err := CheckYourName(nil, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderEnterDateOfBirth, resp.Header.Get("Location"))
}

func TestPostEnterYourNameIsCorrectOnStoreError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{FirstNames: "Bob", LastName: "Smith"}).
		Return(expectedError)

	err := CheckYourName(nil, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterYourNameWithCorrectedName(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"Bobby Smith"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()
	lpa := &page.Lpa{
		Donor:               actor.Donor{Email: "a@example.com"},
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{DeclaredFullName: "Bobby Smith"}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.CertificateProviderNameChangeEmail).
		Return("abc-123")
	notifyClient.
		On("Email", r.Context(), notify.Email{
			EmailAddress:    "a@example.com",
			TemplateID:      "abc-123",
			Personalisation: map[string]string{"declaredName": "Bobby Smith"},
		}).
		Return("", nil)

	err := CheckYourName(nil, donorStore, notifyClient, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderEnterDateOfBirth, resp.Header.Get("Location"))
}

func TestPostEnterYourNameWithCorrectedNameWhenStoreError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {"Bobby Smith"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{DeclaredFullName: "Bobby Smith"}).
		Return(expectedError)

	err := CheckYourName(nil, donorStore, nil, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterYourNameOnValidationError(t *testing.T) {
	form := url.Values{
		"is-name-correct": {"no"},
		"corrected-name":  {""},
	}

	lpa := &page.Lpa{
		CertificateProvider: actor.CertificateProvider{FirstNames: "Bob", LastName: "Smith"},
	}

	data := checkYourNameData{
		App:    testAppData,
		Form:   &checkYourNameForm{IsNameCorrect: "no"},
		Lpa:    lpa,
		Errors: validation.With("corrected-name", validation.EnterError{Label: "yourFullName"}),
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := CheckYourName(template.Execute, donorStore, nil, certificateProviderStore)(testAppData, w, r)
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
