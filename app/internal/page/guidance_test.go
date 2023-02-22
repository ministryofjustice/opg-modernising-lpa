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

	lpa := &Lpa{}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: TestAppData, Lpa: lpa}).
		Return(nil)

	err := Guidance(template.Execute, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenNilDataStore(t *testing.T) {
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

func TestGuidanceWhenDataStoreErrors(t *testing.T) {
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: TestAppData, Lpa: &Lpa{}}).
		Return(ExpectedError)

	err := Guidance(template.Execute, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
}
