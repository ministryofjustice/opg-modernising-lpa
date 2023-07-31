package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCheckYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourLpaData{
			App:  testAppData,
			Form: &checkYourLpaForm{},
			Lpa:  &page.Lpa{},
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		CheckedAndHappy: true,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourLpaData{
			App: testAppData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				CheckedAndHappy: true,
			},
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaDigitalCertificateProviderOnFirstCheck(t *testing.T) {
	testCases := []actor.TaskState{
		actor.TaskNotStarted,
		actor.TaskInProgress,
	}

	for _, existingTaskState := range testCases {
		t.Run(existingTaskState.String(), func(t *testing.T) {
			form := url.Values{
				"checked-and-happy": {"1"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpa := &page.Lpa{
				ID:                  "lpa-id",
				CheckedAndHappy:     false,
				Tasks:               page.Tasks{CheckYourLpa: existingTaskState},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Online},
			}

			updatedLpa := &page.Lpa{
				ID:                  "lpa-id",
				CheckedAndHappy:     true,
				Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Online},
			}

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.
				On("SendCertificateProvider", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, updatedLpa).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), updatedLpa).
				Return(nil)

			err := CheckYourLpa(nil, donorStore, shareCodeSender, nil, nil)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.LpaDetailsSaved.Format("lpa-id")+"?firstCheck=1", resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaDigitalCertificateProviderOnSubsequentChecks(t *testing.T) {
	testCases := map[string]struct {
		certificateProviderDetailsTaskState actor.TaskState
		expectedTemplateId                  notify.Template
		expectedSms                         notify.Sms
	}{
		"cp not started": {
			certificateProviderDetailsTaskState: actor.TaskNotStarted,
			expectedTemplateId:                  notify.CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS,
			expectedSms: notify.Sms{
				PhoneNumber: "07700900000",
				TemplateID:  "template-id",
				Personalisation: map[string]string{
					"donorFullName": "Teneil Throssell",
					"lpaType":       "property and affairs",
				},
			},
		},
		"cp in progress": {
			certificateProviderDetailsTaskState: actor.TaskInProgress,
			expectedTemplateId:                  notify.CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS,
			expectedSms: notify.Sms{
				PhoneNumber: "07700900000",
				TemplateID:  "template-id",
				Personalisation: map[string]string{
					"donorFullNamePossessive": "Teneil Throssell’s",
					"lpaType":                 "property and affairs",
					"lpaId":                   "lpa-id",
					"donorFirstNames":         "Teneil",
				},
			},
		},
		"cp completed": {
			certificateProviderDetailsTaskState: actor.TaskCompleted,
			expectedTemplateId:                  notify.CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS,
			expectedSms: notify.Sms{
				PhoneNumber: "07700900000",
				TemplateID:  "template-id",
				Personalisation: map[string]string{
					"donorFullNamePossessive": "Teneil Throssell’s",
					"lpaType":                 "property and affairs",
					"lpaId":                   "lpa-id",
					"donorFirstNames":         "Teneil",
				},
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
			localizer.
				On("T", "pfaLegalTerm").
				Return("property and affairs")
			localizer.
				On("Possessive", "Teneil Throssell").
				Return("Teneil Throssell’s").
				Maybe()

			testAppData.Localizer = localizer

			lpa := &page.Lpa{
				ID:                  "lpa-id",
				Type:                page.LpaTypePropertyFinance,
				Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAndHappy:     true,
				Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Online, Mobile: "07700900000"},
			}

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", tc.expectedTemplateId).
				Return("template-id")
			notifyClient.
				On("Sms", r.Context(), tc.expectedSms).
				Return("", nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), lpa).
				Return(nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("GetAny", r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{
					Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: tc.certificateProviderDetailsTaskState},
				}, nil)

			err := CheckYourLpa(nil, donorStore, nil, notifyClient, certificateProviderStore)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.LpaDetailsSaved.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaPaperCertificateProviderOnFirstCheck(t *testing.T) {
	testCases := map[actor.TaskState]string{
		actor.TaskNotStarted: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
		actor.TaskInProgress: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
	}

	for existingTaskState, expectedURL := range testCases {
		t.Run(existingTaskState.String(), func(t *testing.T) {
			form := url.Values{
				"checked-and-happy": {"1"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			localizer := newMockLocalizer(t)
			localizer.
				On("T", "pfaLegalTerm").
				Return("property and affairs")

			testAppData.Localizer = localizer

			lpa := &page.Lpa{
				ID:                  "lpa-id",
				Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAndHappy:     false,
				Tasks:               page.Tasks{CheckYourLpa: existingTaskState},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper, Mobile: "07700900000"},
				Type:                page.LpaTypePropertyFinance,
			}

			updatedLpa := &page.Lpa{
				ID:                  "lpa-id",
				Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAndHappy:     true,
				Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper, Mobile: "07700900000"},
				Type:                page.LpaTypePropertyFinance,
			}

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), updatedLpa).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.
				On("TemplateID", notify.CertificateProviderPaperMeetingPromptSMS).
				Return("template-id")
			notifyClient.
				On("Sms", r.Context(), notify.Sms{
					PhoneNumber: "07700900000",
					TemplateID:  "template-id",
					Personalisation: map[string]string{
						"donorFullName":   "Teneil Throssell",
						"lpaType":         "property and affairs",
						"donorFirstNames": "Teneil",
					},
				}).
				Return("", nil)

			err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedURL, resp.Header.Get("Location"))
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

	lpa := &page.Lpa{
		ID:                  "lpa-id",
		Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
		CheckedAndHappy:     true,
		Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
		CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper, Mobile: "07700900000"},
		Type:                page.LpaTypePropertyFinance,
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), lpa).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", notify.CertificateProviderPaperLpaDetailsChangedSMS).
		Return("template-id")
	notifyClient.
		On("Sms", r.Context(), notify.Sms{
			PhoneNumber: "07700900000",
			TemplateID:  "template-id",
			Personalisation: map[string]string{
				"donorFullName":   "Teneil Throssell",
				"lpaId":           "lpa-id",
				"donorFirstNames": "Teneil",
			},
		}).
		Return("", nil)

	err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.LpaDetailsSaved.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCheckYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			CheckedAndHappy: true,
			Tasks:           page.Tasks{CheckYourLpa: actor.TaskCompleted},
		}).
		Return(expectedError)

	err := CheckYourLpa(nil, donorStore, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaWhenShareCodeSenderErrors(t *testing.T) {
	form := url.Values{
		"checked-and-happy": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{
		ID:              "lpa-id",
		CheckedAndHappy: false,
		Tasks:           page.Tasks{CheckYourLpa: actor.TaskInProgress},
	}

	updatedLpa := &page.Lpa{
		ID:              "lpa-id",
		CheckedAndHappy: true,
		Tasks:           page.Tasks{CheckYourLpa: actor.TaskCompleted},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), updatedLpa).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, updatedLpa).
		Return(expectedError)

	err := CheckYourLpa(nil, donorStore, shareCodeSender, nil, nil)(testAppData, w, r, lpa)
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
	localizer.
		On("T", mock.Anything).
		Return("property and affairs")

	testAppData.Localizer = localizer

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.
		On("TemplateID", mock.Anything).
		Return("template-id")
	notifyClient.
		On("Sms", mock.Anything, mock.Anything).
		Return("", expectedError)

	err := CheckYourLpa(nil, donorStore, nil, notifyClient, nil)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{CarryOutBy: actor.Paper}})
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
	template.
		On("Execute", w, mock.MatchedBy(func(data *checkYourLpaData) bool {
			return assert.Equal(t, validation.With("checked-and-happy", validation.SelectError{Label: "theBoxIfYouHaveCheckedAndHappyToShareLpa"}), data.Errors)
		})).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
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
