package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Donor: donor}).
		Return(nil)

	err := Guidance(template.Execute, donorStore)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenNilDataStores(t *testing.T) {
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Execute, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(donor, expectedError)

	err := Guidance(nil, donorStore)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Donor: &actor.DonorProvidedDetails{}}).
		Return(expectedError)

	err := Guidance(template.Execute, donorStore)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}
