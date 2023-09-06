package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{Tasks: page.Tasks{PayForLpa: actor.PaymentTaskDenied}}

	template := newMockTemplate(t)
	template.
		On("Execute", w, feeDeniedData{Lpa: lpa}).
		Return(nil)

	err := FeeDenied(template.Execute, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{Tasks: page.Tasks{PayForLpa: actor.PaymentTaskDenied}}

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, lpa).
		Return(nil)

	err := FeeDenied(nil, payer)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostFeeDeniedWhenPayerError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpa := &page.Lpa{Tasks: page.Tasks{PayForLpa: actor.PaymentTaskDenied}}

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, lpa).
		Return(expectedError)

	err := FeeDenied(nil, payer)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
