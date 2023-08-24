package donor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetResendWitnessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &resendWitnessCodeData{
			App: testAppData,
		}).
		Return(nil)

	err := ResendWitnessCode(template.Execute, nil, time.Now)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetResendWitnessCodeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(template.Execute, nil, time.Now)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostResendWitnessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{
		ID:                    "lpa-id",
		Donor:                 actor.Donor{FirstNames: "john"},
		DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "john"},
	}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("Send", r.Context(), lpa, mock.Anything).
		Return(nil)

	err := ResendWitnessCode(nil, witnessCodeSender, time.Now)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WitnessingAsCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostResendWitnessCodeWhenSendErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{Donor: actor.Donor{FirstNames: "john"}}

	witnessCodeSender := newMockWitnessCodeSender(t)
	witnessCodeSender.
		On("Send", r.Context(), lpa, mock.Anything).
		Return(expectedError)

	err := ResendWitnessCode(nil, witnessCodeSender, time.Now)(testAppData, w, r, lpa)

	assert.Equal(t, expectedError, err)
}

func TestPostResendWitnessCodeWhenTooRecentlySent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{
		Donor:        actor.Donor{FirstNames: "john"},
		WitnessCodes: page.WitnessCodes{{Created: time.Now()}},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &resendWitnessCodeData{
			App:    testAppData,
			Errors: validation.With("request", validation.CustomError{Label: "pleaseWaitOneMinute"}),
		}).
		Return(nil)

	err := ResendWitnessCode(template.Execute, nil, time.Now)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
