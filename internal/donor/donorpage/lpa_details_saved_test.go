package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaDetailsSaved(t *testing.T) {
	testCases := map[string]bool{
		"/?firstCheck=1": true,
		"/":              false,
	}

	for url, expectedIsFirstCheck := range testCases {
		t.Run(url, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, url, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, LpaDetailsSavedData{
					App:          testAppData,
					IsFirstCheck: expectedIsFirstCheck,
					Donor:        &donordata.DonorProvidedDetails{},
				}).
				Return(nil)

			err := LpaDetailsSaved(template.Execute)(testAppData, w, r, &donordata.DonorProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetLpaDetailsSavedOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := LpaDetailsSaved(template.Execute)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
