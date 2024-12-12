package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCheckYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourLpaData{
			App:         testAppData,
			Form:        &checkYourLpaForm{},
			Donor:       &donordata.Provided{},
			CanContinue: true,
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		CheckedAt: testNow,
	}
	donor.UpdateCheckedHash()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourLpaData{
			App:   testAppData,
			Donor: donor,
			Form: &checkYourLpaForm{
				CheckedAndHappy: true,
			},
			CertificateProviderContacted: true,
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaWhenNotChanged(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaID:               "lpa-id",
		CheckedAt:           testNow,
		Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelOnline},
	}
	donor.UpdateCheckedHash()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourLpaData{
			App:   testAppData,
			Donor: donor,
			Form: &checkYourLpaForm{
				CheckedAndHappy: true,
			},
			CertificateProviderContacted: true,
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaDigitalCertificateProviderOnFirstCheck(t *testing.T) {
	testCases := []task.State{
		task.StateNotStarted,
		task.StateInProgress,
	}

	for _, existingTaskState := range testCases {
		t.Run(existingTaskState.String(), func(t *testing.T) {
			form := url.Values{
				"checked-and-happy": {"1"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			uid := actoruid.New()
			provided := &donordata.Provided{
				LpaID:               "lpa-id",
				Hash:                5,
				Tasks:               donordata.Tasks{CheckYourLpa: existingTaskState},
				CertificateProvider: donordata.CertificateProvider{UID: uid, FirstNames: "John", LastName: "Smith", Email: "john@example.com", CarryOutBy: lpadata.ChannelOnline},
			}

			updatedDonor := &donordata.Provided{
				LpaID:                        "lpa-id",
				Hash:                         5,
				CheckedAt:                    testNow,
				Tasks:                        donordata.Tasks{CheckYourLpa: task.StateCompleted},
				CertificateProvider:          donordata.CertificateProvider{UID: uid, FirstNames: "John", LastName: "Smith", Email: "john@example.com", CarryOutBy: lpadata.ChannelOnline},
				CertificateProviderInvitedAt: testNow,
			}
			updatedDonor.UpdateCheckedHash()

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.EXPECT().
				SendCertificateProviderInvite(r.Context(), testAppData, sharecode.CertificateProviderInvite{
					CertificateProviderUID:      provided.CertificateProvider.UID,
					CertificateProviderFullName: provided.CertificateProvider.FullName(),
				}, notify.ToCertificateProvider(provided.CertificateProvider)).
				Return(nil)

			scheduledStore := newMockScheduledStore(t)
			scheduledStore.EXPECT().
				Create(r.Context(), scheduled.Event{
					At:                updatedDonor.CertificateProviderInvitedAt.AddDate(0, 3, 1),
					Action:            scheduled.ActionRemindCertificateProviderToComplete,
					TargetLpaKey:      updatedDonor.PK,
					TargetLpaOwnerKey: updatedDonor.SK,
					LpaUID:            updatedDonor.LpaUID,
				}).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), updatedDonor).
				Return(nil)

			err := CheckYourLpa(nil, donorStore, shareCodeSender, nil, nil, scheduledStore, testNowFn, "http://example.org")(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathLpaDetailsSaved.Format("lpa-id")+"?firstCheck=1", resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaDigitalCertificateProviderOnSubsequentChecks(t *testing.T) {
	testCases := map[string]struct {
		certificateProviderDetailsTaskState task.State
		expectedSms                         notify.SMS
	}{
		"cp not started": {
			certificateProviderDetailsTaskState: task.StateNotStarted,
			expectedSms: notify.CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS{
				DonorFullName: "Teneil Throssell",
				LpaType:       "property and affairs",
			},
		},
		"cp in progress": {
			certificateProviderDetailsTaskState: task.StateInProgress,
			expectedSms: notify.CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS{
				DonorFullNamePossessive: "Teneil Throssell’s",
				LpaType:                 "property and affairs",
				DonorFirstNames:         "Teneil",
			},
		},
		"cp completed": {
			certificateProviderDetailsTaskState: task.StateCompleted,
			expectedSms: notify.CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS{
				DonorFullNamePossessive: "Teneil Throssell’s",
				LpaType:                 "property and affairs",
				DonorFirstNames:         "Teneil",
			},
		},
	}

	for testName, tc := range testCases {
		t.Run(testName, func(t *testing.T) {
			form := url.Values{
				"checked-and-happy": {"1"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("property and affairs")
			localizer.EXPECT().
				Possessive("Teneil Throssell").
				Return("Teneil Throssell’s").
				Maybe()

			testAppData.Localizer = localizer

			provided := &donordata.Provided{
				LpaID:               "lpa-id",
				LpaUID:              "lpa-uid",
				Hash:                5,
				Type:                lpadata.LpaTypePropertyAndAffairs,
				Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAt:           testNow,
				Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
				CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelOnline, Mobile: "07700900000"},
			}

			certificateProvider := &certificateproviderdata.Provided{
				ContactLanguagePreference: localize.Cy,
				Tasks:                     certificateproviderdata.Tasks{ConfirmYourDetails: tc.certificateProviderDetailsTaskState},
			}

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorSMS(r.Context(), notify.ToProvidedCertificateProvider(certificateProvider, provided.CertificateProvider), "lpa-uid", tc.expectedSms).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), provided).
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(r.Context()).
				Return(certificateProvider, nil)

			err := CheckYourLpa(nil, donorStore, nil, notifyClient, certificateProviderStore, nil, testNowFn, "http://example.org")(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathLpaDetailsSaved.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaDigitalCertificateProviderOnSubsequentChecksCertificateProviderStoreErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := CheckYourLpa(nil, nil, nil, nil, certificateProviderStore, nil, testNowFn, "http://example.org")(testAppData, w, r, &donordata.Provided{
		LpaID:               "lpa-id",
		Hash:                5,
		Type:                lpadata.LpaTypePropertyAndAffairs,
		Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
		CheckedAt:           testNow,
		Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelOnline, Mobile: "07700900000"},
	})
	assert.Equal(t, expectedError, err)
}

func TestPostCheckYourLpaPaperCertificateProviderOnFirstCheck(t *testing.T) {
	for _, existingTaskState := range []task.State{task.StateNotStarted, task.StateInProgress} {
		t.Run(existingTaskState.String(), func(t *testing.T) {
			form := url.Values{
				"checked-and-happy": {"1"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("property and affairs")

			testAppData.Localizer = localizer

			provided := &donordata.Provided{
				LpaID:               "lpa-id",
				LpaUID:              "lpa-uid",
				Hash:                5,
				Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				Tasks:               donordata.Tasks{CheckYourLpa: existingTaskState},
				CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper, Mobile: "07700900000"},
				Type:                lpadata.LpaTypePropertyAndAffairs,
			}

			updatedDonor := &donordata.Provided{
				LpaID:               "lpa-id",
				LpaUID:              "lpa-uid",
				Hash:                5,
				Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAt:           testNow,
				Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
				CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper, Mobile: "07700900000"},
				Type:                lpadata.LpaTypePropertyAndAffairs,
			}
			updatedDonor.UpdateCheckedHash()

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), updatedDonor).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorSMS(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), "lpa-uid", notify.CertificateProviderActingOnPaperMeetingPromptSMS{
					DonorFullName:                   "Teneil Throssell",
					LpaType:                         "property and affairs",
					DonorFirstNames:                 "Teneil",
					CertificateProviderStartPageURL: "http://example.org/certificate-provider-start",
				}).
				Return(nil)

			err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathLpaDetailsSaved.Format("lpa-id")+"?firstCheck=1", resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaPaperCertificateProviderOnSubsequentCheck(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:               "lpa-id",
		LpaUID:              "lpa-uid",
		Hash:                5,
		Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
		CheckedAt:           testNow,
		Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper, Mobile: "07700900000"},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), provided).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), "lpa-uid", notify.CertificateProviderActingOnPaperDetailsChangedSMS{
			DonorFullName:   "Teneil Throssell",
			DonorFirstNames: "Teneil",
			LpaUID:          "lpa-uid",
		}).
		Return(nil)

	err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathLpaDetailsSaved.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCheckYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaID:               "lpa-id",
		LpaUID:              "lpa-uid",
		Hash:                5,
		Donor:               donordata.Donor{FirstNames: "Teneil", LastName: "Throssell"},
		CheckedAt:           testNow,
		Tasks:               donordata.Tasks{CheckYourLpa: task.StateCompleted},
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper, Mobile: "07700900000"},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(r.Context(), mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaWhenScheduledStoreErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaID: "lpa-id",
		Hash:  5,
		Tasks: donordata.Tasks{CheckYourLpa: task.StateInProgress},
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(expectedError)

	err := CheckYourLpa(nil, nil, nil, nil, nil, scheduledStore, testNowFn, "http://example.org")(testAppData, w, r, donor)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostCheckYourLpaWhenShareCodeSenderErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaID: "lpa-id",
		Hash:  5,
		Tasks: donordata.Tasks{CheckYourLpa: task.StateInProgress},
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(r.Context(), testAppData, mock.Anything, mock.Anything).
		Return(expectedError)

	err := CheckYourLpa(nil, nil, shareCodeSender, nil, nil, scheduledStore, testNowFn, "http://example.org")(testAppData, w, r, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaWhenNotifyClientErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("property and affairs")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := CheckYourLpa(nil, nil, nil, notifyClient, nil, nil, testNowFn, "http://example.org")(testAppData, w, r, &donordata.Provided{Hash: 5, CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"0"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *checkYourLpaData) bool {
			return assert.Equal(t, validation.With("checked-and-happy", validation.SelectError{Label: "theBoxIfYouHaveCheckedAndHappyToShareLpa"}), data.Errors)
		})).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil, nil, nil, "http://example.org")(testAppData, w, r, &donordata.Provided{Hash: 5})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadCheckYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"checked-and-happy": {" 1   "},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCheckYourLpaForm(r)

	assert.Equal(true, result.CheckedAndHappy)
}

func TestCheckYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *checkYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &checkYourLpaForm{
				CheckedAndHappy: true,
			},
		},
		"invalid": {
			form: &checkYourLpaForm{
				CheckedAndHappy: false,
			},
			errors: validation.
				With("checked-and-happy", validation.SelectError{Label: "theBoxIfYouHaveCheckedAndHappyToShareLpa"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
