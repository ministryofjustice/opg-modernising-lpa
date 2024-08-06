package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestSupporterPathString(t *testing.T) {
	assert.Equal(t, "/supporter/anything", Path("/anything").String())
}

func TestSupporterPathFormat(t *testing.T) {
	assert.Equal(t, "/supporter/anything", Path("/anything").Format())
}

func TestSupporterPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format(), resp.Header.Get("Location"))
}

func TestSupporterPathIsManageOrganisation(t *testing.T) {
	assert.False(t, PathDashboard.IsManageOrganisation())

	assert.True(t, PathOrganisationDetails.IsManageOrganisation())
	assert.True(t, PathEditOrganisationName.IsManageOrganisation())
	assert.True(t, PathManageTeamMembers.IsManageOrganisation())
	assert.True(t, PathEditMember.IsManageOrganisation())
}

func TestSupporterLpaPathString(t *testing.T) {
	assert.Equal(t, "/supporter/anything/{id}", LpaPath("/anything").String())
}

func TestSupporterLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/supporter/anything/abc", LpaPath("/anything").Format("abc"))
}

func TestSupporterLpaPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := LpaPath("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "abc")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("abc"), resp.Header.Get("Location"))
}

func TestSupporterLpaPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := LpaPath("/something")

	err := p.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, "abc", url.Values{"x": {"y"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("abc")+"?x=y", resp.Header.Get("Location"))
}

func TestSupporterLpaPathIsManageOrganisation(t *testing.T) {
	assert.False(t, LpaPath("").IsManageOrganisation())
}
