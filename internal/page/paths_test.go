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
		"redirect": {
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
		"redirect with from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			donor:    &actor.DonorProvidedDetails{LpaID: "lpa-id", Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
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
	assert.True(t, Paths.Supporter.EditOrganisationName.IsManageOrganisation())
	assert.True(t, Paths.Supporter.ManageTeamMembers.IsManageOrganisation())
	assert.True(t, Paths.Supporter.EditMember.IsManageOrganisation())
}

func TestSupporterLpaPathString(t *testing.T) {
	assert.Equal(t, "/supporter/anything/{id}", SupporterLpaPath("/anything").String())
}

func TestSupporterLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/supporter/anything/abc", SupporterLpaPath("/anything").Format("abc"))
}

func TestSupporterLpaPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := SupporterLpaPath("/something")

	err := p.Redirect(w, r, AppData{Lang: localize.En}, "abc")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("abc"), resp.Header.Get("Location"))
}

func TestSupporterLpaPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := SupporterLpaPath("/something")

	err := p.RedirectQuery(w, r, AppData{Lang: localize.En}, "abc", url.Values{"x": {"y"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("abc")+"?x=y", resp.Header.Get("Location"))
}

func TestSupporterLpaPathIsManageOrganisation(t *testing.T) {
	assert.False(t, SupporterLpaPath("").IsManageOrganisation())
}

func TestDonorCanGoTo(t *testing.T) {
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
					AddCorrespondent:           actor.TaskCompleted,
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
					AddCorrespondent:           actor.TaskCompleted,
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
			assert.Equal(t, tc.expected, DonorCanGoTo(tc.donor, tc.url))
		})
	}
}

func TestCertificateProviderCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		certificateProvider *actor.CertificateProviderProvidedDetails
		url                 string
		expected            bool
	}{
		"empty path": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			url:                 "",
			expected:            false,
		},
		"unexpected path": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			url:                 "/whatever",
			expected:            true,
		},
		"unrestricted path": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			url:                 Paths.CertificateProvider.ConfirmYourDetails.Format("123"),
			expected:            true,
		},
		"identity without task": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			url:                 Paths.CertificateProvider.IdentityWithOneLogin.Format("123"),
			expected:            false,
		},
		"identity with task": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.CertificateProvider.IdentityWithOneLogin.Format("123"),
			expected: true,
		},
		"provide certificate without task": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.CertificateProvider.ProvideCertificate.Format("123"),
			expected: false,
		},
		"provide certificate with task": {
			certificateProvider: &actor.CertificateProviderProvidedDetails{
				Tasks: actor.CertificateProviderTasks{
					ConfirmYourDetails:  actor.TaskCompleted,
					ConfirmYourIdentity: actor.TaskCompleted,
				},
			},
			url:      Paths.CertificateProvider.ProvideCertificate.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CertificateProviderCanGoTo(tc.certificateProvider, tc.url))
		})
	}
}

func TestAttorneyCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		attorney *actor.AttorneyProvidedDetails
		url      string
		expected bool
	}{
		"empty path": {
			attorney: &actor.AttorneyProvidedDetails{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			attorney: &actor.AttorneyProvidedDetails{},
			url:      "/whatever",
			expected: true,
		},
		"unrestricted path": {
			attorney: &actor.AttorneyProvidedDetails{},
			url:      Paths.Attorney.ConfirmYourDetails.Format("123"),
			expected: true,
		},
		"sign without task": {
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.Attorney.Sign.Format("123"),
			expected: false,
		},
		"sign with task": {
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			url:      Paths.Attorney.Sign.Format("123"),
			expected: true,
		},
		"would like second signatory not trust corp": {
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
			},
			url:      Paths.Attorney.WouldLikeSecondSignatory.Format("123"),
			expected: false,
		},
		"would like second signatory as trust corp": {
			attorney: &actor.AttorneyProvidedDetails{
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
					ReadTheLpa:         actor.TaskCompleted,
				},
				IsTrustCorporation: true,
			},
			url:      Paths.Attorney.WouldLikeSecondSignatory.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, AttorneyCanGoTo(tc.attorney, tc.url))
		})
	}
}
