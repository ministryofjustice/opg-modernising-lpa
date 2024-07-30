package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSendUsYourEvidenceByPost(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &sendUsYourEvidenceByPostData{App: testAppData}).
		Return(nil)

	err := SendUsYourEvidenceByPost(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSendUsYourEvidenceByPostWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/about-payment", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &sendUsYourEvidenceByPostData{App: testAppData}).
		Return(expectedError)

	err := SendUsYourEvidenceByPost(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSendUsYourEvidenceByPost(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", LpaUID: "lpa-uid", FeeType: pay.HalfFee, EvidenceDelivery: pay.Post}

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(r.Context(), event.ReducedFeeRequested{
			UID:              "lpa-uid",
			RequestType:      pay.HalfFee.String(),
			EvidenceDelivery: pay.Post.String(),
		}).
		Return(nil)

	payer := newMockHandler(t)
	payer.EXPECT().
		Execute(testAppData, w, r, donor).
		Return(nil)

	err := SendUsYourEvidenceByPost(nil, payer.Execute, eventClient)(testAppData, w, r, donor)
	assert.Nil(t, err)
}

func TestPostSendUsYourEvidenceByPostWhenEventClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(r.Context(), mock.Anything).
		Return(expectedError)

	err := SendUsYourEvidenceByPost(nil, nil, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostSendUsYourEvidenceByPostWhenPayerErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/about-payment", nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendReducedFeeRequested(r.Context(), mock.Anything).
		Return(nil)

	payer := newMockHandler(t)
	payer.EXPECT().
		Execute(testAppData, w, r, mock.Anything).
		Return(expectedError)

	err := SendUsYourEvidenceByPost(nil, payer.Execute, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}
