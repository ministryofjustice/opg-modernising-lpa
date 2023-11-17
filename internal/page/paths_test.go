package page

import (
	"net/http"
	"net/http/httptest"
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

func TestLpaPathString(t *testing.T) {
	assert.Equal(t, "/anything", LpaPath("/anything").String())
}

func TestLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/lpa/abc/anything", LpaPath("/anything").Format("abc"))
}

func TestLpaPathRedirect(t *testing.T) {
	testCases := map[string]struct {
		url      string
		lpa      *Lpa
		expected string
	}{
		"allowed": {
			url: "/",
			lpa: &Lpa{
				ID: "lpa-id",
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
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
			lpa:      &Lpa{ID: "lpa-id", Tasks: Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
		},
		"not allowed": {
			url:      "/",
			lpa:      &Lpa{ID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
		"not allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      &Lpa{ID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			Paths.HowToConfirmYourIdentityAndSign.Redirect(w, r, AppData{Lang: localize.En}, tc.lpa)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
		})
	}
}
func TestAttorneyPathString(t *testing.T) {
	assert.Equal(t, "/anything", AttorneyPath("/anything").String())
}

func TestAttorneyPathFormat(t *testing.T) {
	assert.Equal(t, "/attorney/abc/anything", AttorneyPath("/anything").Format("abc"))
}

func TestCertificateProviderPathString(t *testing.T) {
	assert.Equal(t, "/anything", CertificateProviderPath("/anything").String())
}

func TestCertificateProviderPathFormat(t *testing.T) {
	assert.Equal(t, "/certificate-provider/abc/anything", CertificateProviderPath("/anything").Format("abc"))
}
