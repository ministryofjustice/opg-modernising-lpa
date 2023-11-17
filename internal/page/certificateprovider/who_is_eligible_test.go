package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestWhoIsEligible(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.Lpa{Donor: actor.Donor{FirstNames: "Full", LastName: "Name"}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whoIsEligibleData{
			DonorFullName:   "Full Name",
			DonorFirstNames: "Full",
			App:             testAppData,
		}).
		Return(nil)

	err := WhoIsEligible(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWhoIsEligibleWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.Lpa{}, expectedError)

	err := WhoIsEligible(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWhoIsEligibleOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.Lpa{Donor: actor.Donor{FirstNames: "Full", LastName: "Name"}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whoIsEligibleData{
			DonorFullName:   "Full Name",
			DonorFirstNames: "Full",
			App:             testAppData,
		}).
		Return(expectedError)

	err := WhoIsEligible(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
