package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestGetFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, feeDeniedData{Donor: donor, App: testAppData}).
		Return(nil)

	err := FeeDenied(template.Execute, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostFeeDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donor := &actor.DonorProvidedDetails{Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}}

	payer := newMockPayer(t)
	payer.EXPECT().
		Pay(testAppData, w, r, donor).
		Return(nil)

	err := FeeDenied(nil, payer)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostFeeDeniedWhenPayerError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donor := &actor.DonorProvidedDetails{Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}}

	payer := newMockPayer(t)
	payer.EXPECT().
		Pay(testAppData, w, r, donor).
		Return(expectedError)

	err := FeeDenied(nil, payer)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
