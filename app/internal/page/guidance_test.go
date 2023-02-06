package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere", Lpa: lpa}).
		Return(nil)

	err := Guidance(template.Func, "/somewhere", lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGuidanceWhenContinueIsAuthAndLangCy(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	cyAppData := AppData{
		Lang:      Cy,
		SessionID: "session-id",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: cyAppData, Continue: fmt.Sprintf("%s?locale=cy", Paths.Auth), Lpa: lpa}).
		Return(nil)

	err := Guidance(template.Func, Paths.Auth, lpaStore)(cyAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGuidanceWhenNilDataStore(t *testing.T) {
	w := httptest.NewRecorder()

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere"}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Func, "/somewhere", nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGuidanceWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, expectedError)

	err := Guidance(nil, "/somewhere", lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &guidanceData{App: appData, Continue: "/somewhere", Lpa: &Lpa{}}).
		Return(expectedError)

	err := Guidance(template.Func, "/somewhere", lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
