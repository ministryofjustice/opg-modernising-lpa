package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Query: url.Values{}}).
		Return(nil)

	err := Guidance(template.Execute)(testAppData, w, r, nil, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Query: url.Values{}}).
		Return(expectedError)

	err := Guidance(template.Execute)(testAppData, w, r, nil, nil)

	assert.Equal(t, expectedError, err)
}
