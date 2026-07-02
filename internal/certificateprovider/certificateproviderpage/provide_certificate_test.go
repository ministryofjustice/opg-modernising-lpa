package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProvideCertificate(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{
		SignedAt: time.Now(),
		Language: localize.En,
		CertificateProvider: lpadata.CertificateProvider{
			FirstNames: "c",
			LastName:   "d",
		},
		WitnessedByCertificateProviderAt: time.Now(),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &provideCertificateData{
			App:                 testAppData,
			CertificateProvider: &certificateproviderdata.Provided{},
			Lpa:                 donor,
			Form:                newProvideCertificateForm(testAppData.Localizer, localize.En, "c d"),
		}).
		Return(nil)

	err := ProvideCertificate(template.Execute, nil, nil, nil, nil, nil, nil, time.Now, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{}, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideCertificateWhenAlreadyAgreed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{LpaID: "lpa-id", SignedAt: time.Now(), WitnessedByCertificateProviderAt: time.Now()}

	err := ProvideCertificate(nil, nil, nil, nil, nil, nil, nil, time.Now, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{
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

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
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
		SignedAt: testNow,
		Tasks: certificateproviderdata.Tasks{
			ProvideTheCertificate: task.StateCompleted,
		},
		IdentityUserData: identity.UserData{
			Status: identity.StatusConfirmed,
		},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	donor := &donordata.Provided{}
	updatedDonor := &donordata.Provided{
		AttorneysInvitedAt: testNow,
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), certificateProvider).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
			DonorFullNamePossessive:     "c d's",
			DonorFirstNamesPossessive:   "c's",
			LpaType:                     "property-and-affairs",
			CertificateProviderFullName: "a b",
			CertificateProvidedDateTime: "2020-02-03T12:13:14Z",
		}).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, lpa).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(r.Context(), certificateProvider, lpa).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, nil)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(r.Context(), scheduled.Event{
			At:           testNow.AddDate(0, 3, 1),
			Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
			TargetLpaKey: certificateProvider.PK,
			LpaUID:       lpa.LpaUID,
		}, scheduled.Event{
			At:           lpa.SignedAt.AddDate(0, 21, 1),
			Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
			TargetLpaKey: certificateProvider.PK,
			LpaUID:       lpa.LpaUID,
		}).
		Return(nil)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, scheduledStore, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{
		LpaID:                     "lpa-id",
		Email:                     "a@example.com",
		ContactLanguagePreference: localize.En,
		IdentityUserData: identity.UserData{
			Status: identity.StatusConfirmed,
		},
	}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathCertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateWhenIdentityCompleted(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorUID := actoruid.New()
	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d", UID: donorUID, Channel: lpadata.ChannelPaper},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	certificateProvider := &certificateproviderdata.Provided{
		LpaID:    "lpa-id",
		SignedAt: testNow,
		Tasks: certificateproviderdata.Tasks{
			ProvideTheCertificate: task.StateCompleted,
			ConfirmYourIdentity:   task.IdentityStateCompleted,
		},
		IdentityUserData: identity.UserData{
			Status: identity.StatusConfirmed,
		},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	donor := &donordata.Provided{}
	updatedDonor := &donordata.Provided{
		AttorneysInvitedAt: testNow,
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), certificateProvider).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
			DonorFullNamePossessive:     "c d's",
			DonorFirstNamesPossessive:   "c's",
			LpaType:                     "property-and-affairs",
			CertificateProviderFullName: "a b",
			CertificateProvidedDateTime: "2020-02-03T12:13:14Z",
		}).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, lpa).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, nil)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(r.Context(), certificateProvider, lpa).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLetterRequested(r.Context(), event.LetterRequested{
			UID:        "lpa-uid",
			LetterType: "ADVISE_DONOR_CERTIFICATE_HAS_BEEN_PROVIDED",
			ActorType:  actor.TypeDonor,
			ActorUID:   donorUID,
		}).
		Return(nil)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, nil, donorStore, testNowFn, "donorStartURL", eventClient)(testAppData, w, r, &certificateproviderdata.Provided{
		LpaID:                     "lpa-id",
		Email:                     "a@example.com",
		ContactLanguagePreference: localize.En,
		Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
		IdentityUserData: identity.UserData{
			Status: identity.StatusConfirmed,
		},
	}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathCertificateProvided.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateWhenIdentityFailed(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
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
		SignedAt: testNow,
		Tasks: certificateproviderdata.Tasks{
			ProvideTheCertificate: task.StateCompleted,
			ConfirmYourIdentity:   task.IdentityStateCompleted,
		},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	donor := &donordata.Provided{}
	updatedDonor := &donordata.Provided{
		AttorneysInvitedAt: testNow,
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), certificateProvider).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
			DonorFullNamePossessive:     "c d's",
			DonorFirstNamesPossessive:   "c's",
			LpaType:                     "property-and-affairs",
			CertificateProviderFullName: "a b",
			CertificateProvidedDateTime: "2020-02-03T12:13:14Z",
		}).
		Return(nil)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("Dear donor")
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaDonor(lpa), "lpa-uid", notify.CertificateProviderFailedIdentityCheckEmail{
			Greeting:                    "Dear donor",
			CertificateProviderFullName: "a b",
			LpaType:                     "property-and-affairs",
			LpaReferenceNumber:          "lpa-uid",
			DonorStartPageURL:           "donorStartURL",
		}).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, lpa).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, nil)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(r.Context(), certificateProvider, lpa).
		Return(nil)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, nil, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{
		LpaID:                     "lpa-id",
		Email:                     "a@example.com",
		ContactLanguagePreference: localize.En,
		Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
	}, lpa)
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

	signedAt := testNow.Add(-5 * time.Minute)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
			SignedAt:   &signedAt,
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	certificateProvider := &certificateproviderdata.Provided{
		LpaID:                     "lpa-id",
		SignedAt:                  signedAt,
		Tasks:                     certificateproviderdata.Tasks{ProvideTheCertificate: task.StateCompleted},
		IdentityUserData:          identity.UserData{Status: identity.StatusConfirmed},
		ContactLanguagePreference: localize.En,
		Email:                     "a@example.com",
	}

	donor := &donordata.Provided{}
	updatedDonor := &donordata.Provided{
		AttorneysInvitedAt: testNow,
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), certificateProvider).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), "lpa-uid", notify.CertificateProviderCertificateProvidedEmail{
			DonorFullNamePossessive:     "c d's",
			DonorFirstNamesPossessive:   "c's",
			LpaType:                     "property-and-affairs",
			CertificateProviderFullName: "a b",
			CertificateProvidedDateTime: "2020-02-03T12:08:14Z",
		}).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, lpa).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, nil)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(r.Context(), scheduled.Event{
			At:           signedAt.AddDate(0, 3, 1),
			Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
			TargetLpaKey: certificateProvider.PK,
			LpaUID:       lpa.LpaUID,
		}, scheduled.Event{
			At:           lpa.SignedAt.AddDate(0, 21, 1),
			Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
			TargetLpaKey: certificateProvider.PK,
			LpaUID:       lpa.LpaUID,
		}).
		Return(nil)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, nil, scheduledStore, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En, IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, lpa)
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

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow,
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	err := ProvideCertificate(nil, nil, nil, nil, nil, nil, nil, nil, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com"}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathConfirmDontWantToBeCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideCertificateWhenScheduledStoreErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, nil, notifyClient, accessCodeSender, lpaStoreClient, scheduledStore, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En, IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, lpa)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateWhenDonorStoreGetErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := ProvideCertificate(nil, nil, notifyClient, accessCodeSender, lpaStoreClient, nil, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En, IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, lpa)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateWhenDonorStorePutErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow.AddDate(0, -1, 0),
		WitnessedByCertificateProviderAt: testNow,
		CertificateProvider: lpadata.CertificateProvider{
			Email:      "cp@example.org",
			FirstNames: "a",
			LastName:   "b",
		},
		Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:  lpadata.LpaTypePropertyAndAffairs,
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, nil, notifyClient, accessCodeSender, lpaStoreClient, nil, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", Email: "a@example.com", ContactLanguagePreference: localize.En, IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, lpa)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateOnStoreError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, scheduledStore, donorStore, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, &lpadata.Lpa{SignedAt: testNow, WitnessedByCertificateProviderAt: testNow})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateOnEventClientError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLetterRequested(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProvideCertificate(nil, certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, scheduledStore, donorStore, testNowFn, "donorStartURL", eventClient)(testAppData, w, r, &certificateproviderdata.Provided{IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, &lpadata.Lpa{SignedAt: testNow, WitnessedByCertificateProviderAt: testNow, Donor: lpadata.Donor{Channel: lpadata.ChannelPaper}})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateWhenLpaStoreClientError(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &lpadata.Lpa{
		LpaUID:                           "lpa-uid",
		SignedAt:                         testNow,
		WitnessedByCertificateProviderAt: testNow,
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

	err := ProvideCertificate(nil, nil, nil, nil, lpaStoreClient, nil, nil, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, donor)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateOnNotifyClientError(t *testing.T) {
	testcases := map[string]func(*mockNotifyClient){
		"first email": func(notifyClient *mockNotifyClient) {
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)
		},
		"second email": func(notifyClient *mockNotifyClient) {
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				EmailGreeting(mock.Anything).
				Return("")
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	for name, setupNotifyClient := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"agree-to-statement": {"1"},
				"submittable":        {"can-submit"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			notifyClient := newMockNotifyClient(t)
			setupNotifyClient(notifyClient)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			err := ProvideCertificate(nil, nil, notifyClient, nil, lpaStoreClient, nil, nil, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{
				SignedAt:                         testNow,
				WitnessedByCertificateProviderAt: testNow,
				CertificateProvider: lpadata.CertificateProvider{
					Email:      "cp@example.org",
					FirstNames: "a",
					LastName:   "b",
				},
				Donor: lpadata.Donor{FirstNames: "c", LastName: "d"},
				Type:  lpadata.LpaTypePropertyAndAffairs,
			})
			resp := w.Result()

			assert.ErrorIs(t, err, expectedError)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostProvideCertificateWhenAccessCodeSenderErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {"1"},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(r.Context(), testAppData, mock.Anything).
		Return(expectedError)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendCertificateProvider(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := ProvideCertificate(nil, nil, notifyClient, accessCodeSender, lpaStoreClient, nil, nil, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id", IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}}, &lpadata.Lpa{
		SignedAt:                         testNow,
		WitnessedByCertificateProviderAt: testNow,
		Donor:                            lpadata.Donor{FirstNames: "c", LastName: "d"},
		Type:                             lpadata.LpaTypePropertyAndAffairs,
	})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostProvideCertificateWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"agree-to-statement": {""},
		"submittable":        {"can-submit"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *provideCertificateData) bool {
			return assert.Equal(t, []forms.Field{data.Form.AgreeToStatement.Field}, data.Form.Errors) &&
				assert.Equal(t, "errorSelect:Label=toSignAsCertificateProvider", data.Form.AgreeToStatement.Error.Format(testAppData.Localizer))
		})).
		Return(nil)

	err := ProvideCertificate(template.Execute, nil, nil, nil, nil, nil, nil, testNowFn, "donorStartURL", nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{SignedAt: testNow, WitnessedByCertificateProviderAt: testNow})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestToSignCertificateYouMustViewInLanguageError(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("cy").
		Return("Welsh")
	localizer.EXPECT().
		Format("toSignCertificateYouMustViewInLanguage", map[string]any{
			"Lang": "Welsh",
		}).
		Return("some words")

	assert.Equal(t, "some words", toSignCertificateYouMustViewInLanguageError{LpaLanguage: localize.Cy}.Format(localizer))
}
