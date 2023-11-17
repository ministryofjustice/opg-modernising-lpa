package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWouldYouLikeToSendEvidence(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howWouldYouLikeToSendEvidenceData{
			App:     testAppData,
			Options: pay.EvidenceDeliveryValues,
		}).
		Return(nil)

	err := HowWouldYouLikeToSendEvidence(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldYouLikeToSendEvidenceFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &howWouldYouLikeToSendEvidenceData{
			App:     testAppData,
			Options: pay.EvidenceDeliveryValues,
		}).
		Return(nil)

	err := HowWouldYouLikeToSendEvidence(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldYouLikeToSendEvidenceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := HowWouldYouLikeToSendEvidence(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldYouLikeToSendEvidence(t *testing.T) {
	testcases := map[pay.EvidenceDelivery]page.LpaPath{
		pay.Upload: page.Paths.UploadEvidence,
		pay.Post:   page.Paths.SendUsYourEvidenceByPost,
	}

	for evidenceDelivery, redirect := range testcases {
		t.Run(evidenceDelivery.String(), func(t *testing.T) {
			form := url.Values{
				"evidence-delivery": {evidenceDelivery.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &actor.DonorProvidedDetails{ID: "lpa-id", EvidenceDelivery: evidenceDelivery}).
				Return(nil)

			err := HowWouldYouLikeToSendEvidence(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWouldYouLikeToSendEvidenceWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"evidence-delivery": {pay.Upload.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWouldYouLikeToSendEvidence(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{ID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostHowWouldYouLikeToSendEvidenceWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *howWouldYouLikeToSendEvidenceData) bool {
			return assert.Equal(t, validation.With("evidence-delivery", validation.SelectError{Label: "howYouWouldLikeToSendUsYourEvidence"}), data.Errors)
		})).
		Return(nil)

	err := HowWouldYouLikeToSendEvidence(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
