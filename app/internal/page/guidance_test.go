package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: TestAppData}).
		Return(nil)

	err := Guidance(template.Execute, nil)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenNilDataStores(t *testing.T) {
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: TestAppData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Execute, nil)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, ExpectedError)

	err := Guidance(nil, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: TestAppData}).
		Return(ExpectedError)

	err := Guidance(template.Execute, nil)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
}
