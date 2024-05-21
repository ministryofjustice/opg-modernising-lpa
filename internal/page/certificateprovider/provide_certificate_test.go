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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProvideCertificate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{SignedAt: time.Now()}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &provideCertificateData{
			App:                 testAppData,
			CertificateProvider: &actor.CertificateProviderProvidedDetails{},
			Lpa:                 donor,
			Form:                &provideCertificateForm{},
		}).
		Return(nil)

	err := ProvideCertificate(template.Execute, lpaStoreResolvingService, certificateProviderStore, nil, nil, nil, time.Now)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideCertificateRedirectsToStartOnLpaNotSubmitted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaID: "lpa-id"}, nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, nil, nil, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetProvideCertificateWhenAlreadyAgreed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{LpaID: "lpa-id", SignedAt: time.Now()}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{
			Certificate: actor.Certificate{
				Agreed: time.Now(),
			},
		}, nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, nil, nil, nil, time.Now)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.CertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetProvideCertificateWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := ProvideCertificate(nil, lpaStoreResolvingService, nil, nil, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificate(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpa := &lpastore.Lpa{
		LpaUID:   "lpa-uid",
		SignedAt: now,
		CertificateProvider: lpastore.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: actor.Donor{FirstNames: "c", LastName: "d"},
		Type:  actor.LpaTypePropertyAndAffairs,
	}

	certificateProvider := &actor.CertificateProviderProvidedDetails{
		LpaID: "lpa-id",
		Certificate: actor.Certificate{
			AgreeToStatement: true,
			Agreed:           now,
		},
		Tasks: actor.CertificateProviderTasks{
			ProvideTheCertificate: actor.TaskCompleted,
		},
		Email: "a@example.com",
	}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", Email: "a@example.com"}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), certificateProvider).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive("c").
		Return("the possessive first names")
	localizer.EXPECT().
		Possessive("c d").
		Return("the possessive full name")
	localizer.EXPECT().
		T("property-and-affairs").
		Return("the translated term")
	localizer.EXPECT().
		FormatDateTime(now).
		Return("the formatted date")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), "a@example.com", "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
			DonorFullNamePossessive:     "the possessive full name",
			DonorFirstNamesPossessive:   "the possessive first names",
			LpaType:                     "the translated term",
			CertificateProviderFullName: "a b",
			CertificateProvidedDateTime: "the formatted date",
		}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, lpa).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(r.Context(), certificateProvider, lpa).
		Return(nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.CertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateWhenCannotSubmit(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"cannot-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpa := &lpastore.Lpa{
		LpaUID:   "lpa-uid",
		SignedAt: now,
		CertificateProvider: lpastore.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: actor.Donor{FirstNames: "c", LastName: "d"},
		Type:  actor.LpaTypePropertyAndAffairs,
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", Email: "a@example.com"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.ConfirmDontWantToBeCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateOnStoreError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{SignedAt: now}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("")
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("")
	localizer.EXPECT().
		T(mock.Anything).
		Return("")
	localizer.EXPECT().
		FormatDateTime(mock.Anything).
		Return("")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, mock.Anything).
		Return(nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificateWhenLpaStoreClientError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donor := &lpastore.Lpa{
		LpaUID:   "lpa-uid",
		SignedAt: now,
		CertificateProvider: lpastore.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: actor.Donor{FirstNames: "c", LastName: "d"},
		Type:  actor.LpaTypePropertyAndAffairs,
	}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, nil, nil, lpaStoreClient, func() time.Time { return now })(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostProvideCertificateOnNotifyClientError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			SignedAt: now,
			CertificateProvider: lpastore.CertificateProvider{
				Email:      "cp@example.org",
				FirstNames: "a",
				LastName:   "b",
			},
			Donor: actor.Donor{FirstNames: "c", LastName: "d"},
			Type:  actor.LpaTypePropertyAndAffairs,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("")
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("")
	localizer.EXPECT().
		T(mock.Anything).
		Return("")
	localizer.EXPECT().
		FormatDateTime(mock.Anything).
		Return("")

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	testAppData.Localizer = localizer

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, notifyClient, nil, lpaStoreClient, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, fmt.Errorf("email failed: %w", expectedError), err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideCertificateWhenShareCodeSenderErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{
			SignedAt: now,
			Donor:    actor.Donor{FirstNames: "c", LastName: "d"},
			Type:     actor.LpaTypePropertyAndAffairs,
		}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive("c").
		Return("the possessive first names")
	localizer.EXPECT().
		Possessive("c d").
		Return("the possessive full name")
	localizer.EXPECT().
		T("property-and-affairs").
		Return("the translated term")
	localizer.EXPECT().
		FormatDateTime(now).
		Return("the formatted date")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, mock.Anything).
		Return(expectedError)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := ProvideCertificate(nil, lpaStoreResolvingService, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostProvideCertificateWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {""},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{SignedAt: now}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *provideCertificateData) bool {
			return assert.Equal(t, validation.With("agree-to-statement", validation.SelectError{Label: "toSignAsCertificateProvider"}), data.Errors)
		})).
		Return(nil)

	err := ProvideCertificate(template.Execute, lpaStoreResolvingService, certificateProviderStore, nil, nil, nil, func() time.Time { return now })(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadProvideCertificateForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"agree-to-statement": {" 1   "},
		"submittable":        {"can-submit"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readProvideCertificateForm(r)

	assert.Equal(true, result.AgreeToStatement)
	assert.Equal("can-submit", result.Submittable)
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
