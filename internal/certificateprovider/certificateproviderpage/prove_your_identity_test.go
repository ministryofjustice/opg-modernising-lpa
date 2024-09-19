package certificateproviderpage

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmYourIdentity(t *testing.T) {
	testcases := []bool{
		true, false,
	}

	for _, enabled := range testcases {
		t.Run(fmt.Sprintf("enabled=%t", enabled), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &proveYourIdentityData{
					App:                  testAppData,
					Errors:               nil,
					LowConfidenceEnabled: enabled,
				}).
				Return(nil)

			err := ProveYourIdentity(template.Execute, enabled)(testAppData, w, r, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestConfirmYourIdentityTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ProveYourIdentity(template.Execute, true)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
