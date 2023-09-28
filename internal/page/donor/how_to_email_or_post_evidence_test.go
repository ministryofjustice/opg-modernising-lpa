package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowToEmailOrPostEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howToEmailOrPostEvidenceData{App: testAppData}).
		Return(nil)

	err := HowToEmailOrPostEvidence(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowToEmailOrPostEvidenceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howToEmailOrPostEvidenceData{App: testAppData}).
		Return(expectedError)

	err := HowToEmailOrPostEvidence(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowToEmailOrPostEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	lpa := &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}}

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, lpa).
		Return(nil)

	err := HowToEmailOrPostEvidence(nil, payer)(testAppData, w, r, lpa)
	assert.Nil(t, err)
}

func TestPostHowToEmailOrPostEvidenceWhenPayerErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, mock.Anything).
		Return(expectedError)

	err := HowToEmailOrPostEvidence(nil, payer)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}
