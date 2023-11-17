package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestPathString(t *testing.T) {
	assert.Equal(t, "/anything", Path("/anything").String())
}

func TestPathFormat(t *testing.T) {
	assert.Equal(t, "/anything", Path("/anything").Format())
}

func TestPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, AppData{Lang: localize.En})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format(), resp.Header.Get("Location"))
}

func TestPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.RedirectQuery(w, r, AppData{Lang: localize.En}, url.Values{"q": {"1"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format()+"?q=1", resp.Header.Get("Location"))
}

func TestLpaPathString(t *testing.T) {
	assert.Equal(t, "/anything", LpaPath("/anything").String())
}

func TestLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/lpa/abc/anything", LpaPath("/anything").Format("abc"))
}

func TestLpaPathRedirect(t *testing.T) {
	testCases := map[string]struct {
		url      string
		lpa      *actor.DonorProvidedDetails
		expected string
	}{
		"allowed": {
			url: "/",
			lpa: &actor.DonorProvidedDetails{
				ID: "lpa-id",
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypeHealthWelfare,
				Tasks: actor.DonorTasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
					PayForLpa:                  actor.PaymentTaskCompleted,
				},
			},
			expected: Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id"),
		},
		"allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      &actor.DonorProvidedDetails{ID: "lpa-id", Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
		},
		"not allowed": {
			url:      "/",
			lpa:      &actor.DonorProvidedDetails{ID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
		"not allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      &actor.DonorProvidedDetails{ID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := Paths.HowToConfirmYourIdentityAndSign.Redirect(w, r, AppData{Lang: localize.En}, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
		})
	}
}

func TestLpaPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	err := Paths.TaskList.RedirectQuery(w, r, AppData{Lang: localize.En}, &actor.DonorProvidedDetails{ID: "lpa-id"}, url.Values{"q": {"1"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.TaskList.Format("lpa-id")+"?q=1", resp.Header.Get("Location"))
}

func TestAttorneyPathString(t *testing.T) {
	assert.Equal(t, "/anything", AttorneyPath("/anything").String())
}

func TestAttorneyPathFormat(t *testing.T) {
	assert.Equal(t, "/attorney/abc/anything", AttorneyPath("/anything").Format("abc"))
}

func TestAttorneyPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := AttorneyPath("/something")

	err := p.Redirect(w, r, AppData{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestAttorneyPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := AttorneyPath("/something")

	err := p.RedirectQuery(w, r, AppData{Lang: localize.En}, "lpa-id", url.Values{"q": {"1"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id")+"?q=1", resp.Header.Get("Location"))
}

func TestCertificateProviderPathString(t *testing.T) {
	assert.Equal(t, "/anything", CertificateProviderPath("/anything").String())
}

func TestCertificateProviderPathFormat(t *testing.T) {
	assert.Equal(t, "/certificate-provider/abc/anything", CertificateProviderPath("/anything").Format("abc"))
}

func TestCertificateProviderPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := CertificateProviderPath("/something")

	err := p.Redirect(w, r, AppData{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id"), resp.Header.Get("Location"))
}
