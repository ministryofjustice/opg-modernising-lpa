package page

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTemplate struct {
	mock.Mock
}

func (t *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := t.Called(w, data)
	return args.Error(0)
}

func TestStart(t *testing.T) {
	w := httptest.NewRecorder()

	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, startData{L: localizer, Lang: En}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	Start(nil, localizer, En, template.Func)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}
