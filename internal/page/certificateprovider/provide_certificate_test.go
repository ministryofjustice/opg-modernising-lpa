package certificateprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProvideCertificate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{SignedAt: time.Now()}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &provideCertificateData{
			App:                 testAppData,
			CertificateProvider: &actor.CertificateProviderProvidedDetails{},
			Donor:               donor,
			Form:                &provideCertificateForm{},
		}).
		Return(nil)

	err := ProvideCertificate(template.Execute, donorStore, time.Now, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideCertificateRedirectsToStartOnLpaNotSubmitted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{LpaID: "lpa-id"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	err := ProvideCertificate(nil, donorStore, nil, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetProvideCertificateWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := ProvideCertificate(nil, donorStore, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideCertificateWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := ProvideCertificate(nil, donorStore, nil, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificate(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{
			SignedAt: now,
			CertificateProvider: actor.CertificateProvider{
				Email:      "cp@example.org",
				FirstNames: "a",
				LastName:   "b",
			},
			Donor: actor.Donor{FirstNames: "c", LastName: "d"},
			Type:  actor.LpaTypePropertyFinance,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{
			LpaID: "lpa-id",
			Certificate: actor.Certificate{
				AgreeToStatement: true,
				Agreed:           now,
			},
			Tasks: actor.CertificateProviderTasks{
				ProvideTheCertificate: actor.TaskCompleted,
			},
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.
		On("Possessive", "c").
		Return("the possessive first names")
	localizer.
		On("Possessive", "c d").
		Return("the possessive full name")
	localizer.
		On("T", "pfaLegalTerm").
		Return("the translated term")
	localizer.
		On("FormatDateTime", now).
		Return("the formatted date")

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("Email", r.Context(), notify.Email{
			EmailAddress: "cp@example.org",
			TemplateID:   "the-template-id",
			Personalisation: map[string]string{
				"donorFullNamePossessive":     "the possessive full name",
				"donorFirstNamesPossessive":   "the possessive first names",
				"lpaLegalTerm":                "the translated term",
				"certificateProviderFullName": "a b",
				"certificateProvidedDateTime": "the formatted date",
			},
		}).
		Return("", nil)
	notifyClient.
		On("TemplateID", notify.CertificateProviderCertificateProvidedEmail).
		Return("the-template-id")

	testAppData.Localizer = localizer

	err := ProvideCertificate(nil, donorStore, func() time.Time { return now }, certificateProviderStore, notifyClient)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.CertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateOnStoreError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{SignedAt: now}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, donorStore, func() time.Time { return now }, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificateOnNoticyClientError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{
			SignedAt: now,
			CertificateProvider: actor.CertificateProvider{
				Email:      "cp@example.org",
				FirstNames: "a",
				LastName:   "b",
			},
			Donor: actor.Donor{FirstNames: "c", LastName: "d"},
			Type:  actor.LpaTypePropertyFinance,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.
		On("Possessive", mock.Anything).
		Return("")
	localizer.
		On("Possessive", mock.Anything).
		Return("")
	localizer.
		On("T", mock.Anything).
		Return("")
	localizer.
		On("FormatDateTime", mock.Anything).
		Return("")

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("Email", r.Context(), mock.Anything).
		Return("", expectedError)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("")

	testAppData.Localizer = localizer

	err := ProvideCertificate(nil, donorStore, func() time.Time { return now }, certificateProviderStore, notifyClient)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, fmt.Errorf("email failed: %w", expectedError), err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificateWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{SignedAt: now}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *provideCertificateData) bool {
			return assert.Equal(t, validation.With("agree-to-statement", validation.SelectError{Label: "toSignAsCertificateProvider"}), data.Errors)
		})).
		Return(nil)

	err := ProvideCertificate(template.Execute, donorStore, func() time.Time { return now }, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadProvideCertificateForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"agree-to-statement": {" 1   "},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readProvideCertificateForm(r)

	assert.Equal(true, result.AgreeToStatement)
}

func TestProvideCertificateFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *provideCertificateForm
		errors validation.List
	}{
		"valid": {
			form: &provideCertificateForm{
				AgreeToStatement: true,
			},
		},
		"invalid": {
			form: &provideCertificateForm{},
			errors: validation.
				With("agree-to-statement", validation.SelectError{Label: "toSignAsCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
