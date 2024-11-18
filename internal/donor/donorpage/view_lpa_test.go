package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestViewLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{LpaUID: "M-0000"}
	lpa := &lpadata.Lpa{LpaUID: "M-0000"}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "M-0000").
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLpaData{App: testAppData, Lpa: lpa}).
		Return(nil)

	err := ViewLpa(template.Execute, lpaStoreClient)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestViewLpaWhenNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{LpaUID: "M-0000"}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "M-0000").
		Return(nil, lpastore.ErrNotFound)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLpaData{App: testAppData, Lpa: lpastore.FromDonorProvidedDetails(donor)}).
		Return(nil)

	err := ViewLpa(template.Execute, lpaStoreClient)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestViewLpaWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := ViewLpa(nil, lpaStoreClient)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestViewLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ViewLpa(template.Execute, lpaStoreClient)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}
