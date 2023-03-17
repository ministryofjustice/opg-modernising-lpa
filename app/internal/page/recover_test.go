package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecover(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ContextWithAppData(context.Background(), TestAppData), http.MethodGet, "/", nil)

	badHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uninitialised http.Handler
		uninitialised.ServeHTTP(w, r)
	})

	template := newMockTemplate(t)
	template.
		On("Execute", w, &errorData{App: TestAppData}).
		Return(nil)

	logger := newMockLogger(t)
	logger.
		On("Request", r, mock.MatchedBy(func(e recoverError) bool {
			return assert.Equal(t, "recover error", e.Error()) &&
				assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", e.Title()) &&
				assert.Contains(t, e.Data(), "github.com/ministryofjustice/opg-modernising-lpa/internal/page.TestRecover") &&
				assert.Contains(t, e.Data(), "app/internal/page/recover_test.go:")
		}))

	Recover(template.Execute, logger, badHandler).ServeHTTP(w, r)
}
