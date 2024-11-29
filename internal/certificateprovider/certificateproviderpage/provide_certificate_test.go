package certificateproviderpage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProvideCertificate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{SignedAt: time.Now(), WitnessedByCertificateProviderAt: time.Now()}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &provideCertificateData{
			App:                 testAppData,
			CertificateProvider: &certificateproviderdata.Provided{},
			Lpa:                 donor,
			Form:                &provideCertificateForm{},
		}).
		Return(nil)

	err := ProvideCertificate(template.Execute, nil, nil, nil, nil, time.Now)(testAppData, w, r, &certificateproviderdata.Provided{}, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideCertificateRedirectsToStartOnLpaNotSubmitted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ProvideCertificate(nil, nil, nil, nil, nil, nil)(testAppData, w, r, nil, &lpadata.Lpa{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetProvideCertificateWhenAlreadyAgreed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{LpaID: "lpa-id", SignedAt: time.Now(), WitnessedByCertificateProviderAt: time.Now()}

	err := ProvideCertificate(nil, nil, nil, nil, nil, time.Now)(testAppData, w, r, &certificateproviderdata.Provided{
		SignedAt: time.Now(),
	}, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathCertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
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

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	certificateProvider := &certificateproviderdata.Provided{
		LpaID:    "lpa-id",
		SignedAt: now,
		Tasks: certificateproviderdata.Tasks{
			ProvideTheCertificate: task.StateCompleted,
		},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
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
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
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

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathCertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateWhenSignedInLpaStore(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	signedAt := time.Now().Add(-5 * time.Minute)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
			SignedAt:   signedAt,
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	certificateProvider := &certificateproviderdata.Provided{
		LpaID:    "lpa-id",
		SignedAt: signedAt,
		Tasks: certificateproviderdata.Tasks{
			ProvideTheCertificate: task.StateCompleted,
		},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
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
		FormatDateTime(signedAt).
		Return("the formatted date")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
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

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, shareCodeSender, nil, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathCertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
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

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	err := ProvideCertificate(nil, nil, nil, nil, nil, nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com"}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathConfirmDontWantToBeCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
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

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
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

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{SignedAt: now, WitnessedByCertificateProviderAt: now})
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

	donor := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, nil, nil, nil, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, donor)
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

	err := ProvideCertificate(nil, nil, notifyClient, nil, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	})
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

	err := ProvideCertificate(nil, nil, notifyClient, shareCodeSender, lpaStoreClient, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{
		SignedAt:                         now,
		WitnessedByCertificateProviderAt: now,
		Donor:                            lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:                             lpadata.LpaTypePropertyAndAffairs,
	})
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

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *provideCertificateData) bool {
			return assert.Equal(t, validation.With("agree-to-statement", validation.SelectError{Label: "toSignAsCertificateProvider"}), data.Errors)
		})).
		Return(nil)

	err := ProvideCertificate(template.Execute, nil, nil, nil, nil, func() time.Time { return now })(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{SignedAt: now, WitnessedByCertificateProviderAt: now})
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
