package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDeleteLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &deleteLpaData{
			App:   testAppData,
			Donor: &donordata.DonorProvidedDetails{},
		}).
		Return(nil)

	err := DeleteLpa(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DeleteLpa(template.Execute, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDeleteLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Delete(r.Context()).
		Return(nil)

	err := DeleteLpa(nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaUID: "lpa-uid"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.LpaDeleted.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostDeleteLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Delete(r.Context()).
		Return(expectedError)

	err := DeleteLpa(nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}
