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

	err := CheckYourLpa(template.Execute, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
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

	err := CheckYourLpa(template.Execute, nil, nil, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpaDigitalCertificateProvider(t *testing.T) {
	testCases := map[actor.TaskState]string{
		actor.TaskNotStarted: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
		actor.TaskInProgress: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
		actor.TaskCompleted:  page.Paths.LpaDetailsSaved.Format("lpa-id"),
	}

	for existingTaskState, expectedURL := range testCases {
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
				CertificateProvider: actor.CertificateProvider{CarryOutBy: "online"},
			}

			updatedLpa := &page.Lpa{
				ID:                  "lpa-id",
				CheckedAndHappy:     true,
				Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: "online"},
			}

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), updatedLpa).
				Return(nil)

			shareCodeSender := newMockShareCodeSender(t)
			shareCodeSender.
				On("SendCertificateProvider", r.Context(), notify.CertificateProviderInviteEmail, testAppData, true, updatedLpa).
				Return(nil)

			err := CheckYourLpa(nil, donorStore, shareCodeSender, nil)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedURL, resp.Header.Get("Location"))
		})
	}
}

func TestPostCheckYourLpaPaperCertificateProvider(t *testing.T) {
	testCases := map[actor.TaskState]string{
		actor.TaskNotStarted: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
		actor.TaskInProgress: page.Paths.LpaDetailsSaved.Format("lpa-id") + "?firstCheck=1",
		actor.TaskCompleted:  page.Paths.LpaDetailsSaved.Format("lpa-id"),
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
				CertificateProvider: actor.CertificateProvider{CarryOutBy: "paper", Mobile: "07700900000"},
				Type:                page.LpaTypePropertyFinance,
			}

			updatedLpa := &page.Lpa{
				ID:                  "lpa-id",
				Donor:               actor.Donor{FirstNames: "Teneil", LastName: "Throssell"},
				CheckedAndHappy:     true,
				Tasks:               page.Tasks{CheckYourLpa: actor.TaskCompleted},
				CertificateProvider: actor.CertificateProvider{CarryOutBy: "paper", Mobile: "07700900000"},
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

			err := CheckYourLpa(nil, donorStore, nil, notifyClient)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedURL, resp.Header.Get("Location"))
		})
	}
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

	err := CheckYourLpa(nil, donorStore, nil, nil)(testAppData, w, r, &page.Lpa{})
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

	err := CheckYourLpa(nil, donorStore, shareCodeSender, nil)(testAppData, w, r, lpa)
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

	err := CheckYourLpa(nil, donorStore, nil, notifyClient)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{CarryOutBy: "paper"}})
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

	err := CheckYourLpa(template.Execute, nil, nil, nil)(testAppData, w, r, &page.Lpa{})
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
