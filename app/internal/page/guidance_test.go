package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &guidanceData{App: TestAppData, Continue: "/somewhere", Lpa: lpa}).
		Return(nil)

	err := Guidance(template.Func, "/somewhere", lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGuidanceWhenContinueIsAuthAndLangCy(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	cyAppData := AppData{
		Lang:      localize.Cy,
		SessionID: "session-id",
	}

	template := &MockTemplate{}
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

	template := &MockTemplate{}
	template.
		On("Func", w, &guidanceData{App: TestAppData, Continue: "/somewhere"}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Func, "/somewhere", nil)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGuidanceWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, ExpectedError)

	err := Guidance(nil, "/somewhere", lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &guidanceData{App: TestAppData, Continue: "/somewhere", Lpa: &Lpa{}}).
		Return(ExpectedError)

	err := Guidance(template.Func, "/somewhere", lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
