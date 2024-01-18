package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?a=b", nil)

	donor := &actor.DonorProvidedDetails{}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Donor: donor, Query: url.Values{"a": {"b"}}}).
		Return(nil)

	err := Guidance(template.Execute)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Donor: &actor.DonorProvidedDetails{}, Query: url.Values{}}).
		Return(expectedError)

	err := Guidance(template.Execute)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}
