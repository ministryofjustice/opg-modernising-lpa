package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestYouHaveDecidedNotToBeACertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?donorFullName=a+b+c", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, youHaveDecidedNotToBeACertificateProviderData{
			App:           testAppData,
			DonorFullName: "a b c",
		}).
		Return(nil)

	err := YouHaveDecidedNotToBeACertificateProvider(template.Execute)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestYouHaveDecidedNotToBeACertificateProviderWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?donorFullName=a+b+c", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := YouHaveDecidedNotToBeACertificateProvider(template.Execute)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
