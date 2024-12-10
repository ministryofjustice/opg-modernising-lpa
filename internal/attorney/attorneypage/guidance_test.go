package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Lpa: lpa}).
		Return(nil)

	err := Guidance(template.Execute)(testAppData, w, r, nil, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Lpa: &lpadata.Lpa{}}).
		Return(expectedError)

	err := Guidance(template.Execute)(testAppData, w, r, nil, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}
