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
	assert.Equal(t, "/lpa/{id}/anything", LpaPath("/anything").String())
}

func TestLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/lpa/abc/anything", LpaPath("/anything").Format("abc"))
}

func TestLpaPathRedirect(t *testing.T) {
	testCases := map[string]struct {
		url      string
		donor    *actor.DonorProvidedDetails
		expected string
	}{
		"allowed": {
			url: "/",
			donor: &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypePersonalWelfare,
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
			donor:    &actor.DonorProvidedDetails{LpaID: "lpa-id", Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
		},
		"not allowed": {
			url:      "/",
			donor:    &actor.DonorProvidedDetails{LpaID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
		"not allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			donor:    &actor.DonorProvidedDetails{LpaID: "lpa-id"},
			expected: Paths.TaskList.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := Paths.HowToConfirmYourIdentityAndSign.Redirect(w, r, AppData{Lang: localize.En}, tc.donor)
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

	err := Paths.TaskList.RedirectQuery(w, r, AppData{Lang: localize.En}, &actor.DonorProvidedDetails{LpaID: "lpa-id"}, url.Values{"q": {"1"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, Paths.TaskList.Format("lpa-id")+"?q=1", resp.Header.Get("Location"))
}

func TestAttorneyPathString(t *testing.T) {
	assert.Equal(t, "/attorney/{id}/anything", AttorneyPath("/anything").String())
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
	assert.Equal(t, "/certificate-provider/{id}/anything", CertificateProviderPath("/anything").String())
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

func TestSupporterPathString(t *testing.T) {
	assert.Equal(t, "/supporter/anything", SupporterPath("/anything").String())
}

func TestSupporterPathFormat(t *testing.T) {
	assert.Equal(t, "/supporter/anything", SupporterPath("/anything").Format())
}

func TestSupporterPathFormatID(t *testing.T) {
	testcases := []struct {
		supporterPath SupporterPath
		expectedPath  string
	}{
		{
			supporterPath: SupporterPath("/anything/{id}"),
			expectedPath:  "/supporter/anything/1",
		},
		{
			supporterPath: SupporterPath("/{id}/anything"),
			expectedPath:  "/supporter/1/anything",
		},
		{
			supporterPath: SupporterPath("/{id}/anything/{id}"),
			expectedPath:  "/supporter/1/anything/{id}",
		},
	}

	for _, tc := range testcases {
		assert.Equal(t, tc.expectedPath, tc.supporterPath.FormatID("1"))
	}

}

func TestSupporterPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := SupporterPath("/something")

	err := p.Redirect(w, r, AppData{Lang: localize.En})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format(), resp.Header.Get("Location"))
}

func TestSupporterPathIsManageOrganisation(t *testing.T) {
	assert.False(t, Paths.Supporter.Dashboard.IsManageOrganisation())
	assert.True(t, Paths.Supporter.OrganisationDetails.IsManageOrganisation())
}

func TestCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		donor    *actor.DonorProvidedDetails
		url      string
		expected bool
	}{
		"empty path": {
			donor:    &actor.DonorProvidedDetails{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			donor:    &actor.DonorProvidedDetails{},
			url:      "/whatever",
			expected: true,
		},
		"getting help signing no certificate provider": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					YourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: false,
		},
		"getting help signing": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					CertificateProvider: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: true,
		},
		"check your lpa when unsure if can sign": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: false,
		},
		"check your lpa when can sign": {
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{CanSign: form.Yes},
				Type:  actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: true,
		},
		"about payment without task": {
			donor:    &actor.DonorProvidedDetails{LpaID: "123"},
			url:      Paths.AboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypePropertyAndAffairs,
				Tasks: actor.DonorTasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					WhenCanTheLpaBeUsed:        actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
				},
			},
			url:      Paths.AboutPayment.Format("123"),
			expected: true,
		},
		"identity without task": {
			donor:    &actor.DonorProvidedDetails{},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: false,
		},
		"identity with tasks": {
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypePersonalWelfare,
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
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CanGoTo(tc.donor, tc.url))
		})
	}
}
