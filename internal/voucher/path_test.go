package voucher

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestPathString(t *testing.T) {
	assert.Equal(t, "/voucher/{id}/anything", Path("/anything").String())
}

func TestPathFormat(t *testing.T) {
	assert.Equal(t, "/voucher/abc/anything", Path("/anything").Format("abc"))
}

func TestPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPathRedirectWhenFrom(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/?from=/x", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/x", resp.Header.Get("Location"))
}
