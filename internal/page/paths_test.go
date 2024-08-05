package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format(), resp.Header.Get("Location"))
}

func TestPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, url.Values{"q": {"1"}})
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
		donor    *donordata.Provided
		expected string
	}{
		"redirect": {
			url: "/",
			donor: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePersonalWelfare,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					LifeSustainingTreatment:    task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					CheckYourLpa:               task.StateCompleted,
					PayForLpa:                  task.PaymentStateCompleted,
				},
			},
			expected: Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id"),
		},
		"redirect with from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			donor:    &donordata.Provided{LpaID: "lpa-id", Tasks: donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := Paths.HowToConfirmYourIdentityAndSign.Redirect(w, r, appcontext.Data{Lang: localize.En}, tc.donor)
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

	err := Paths.TaskList.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, &donordata.Provided{LpaID: "lpa-id"}, url.Values{"q": {"1"}})
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

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestAttorneyPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := AttorneyPath("/something")

	err := p.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, "lpa-id", url.Values{"q": {"1"}})
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

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
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

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En})
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

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "abc")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("abc"), resp.Header.Get("Location"))
}

func TestSupporterLpaPathRedirectQuery(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := SupporterLpaPath("/something")

	err := p.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, "abc", url.Values{"x": {"y"}})
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
		donor    *donordata.Provided
		url      string
		expected bool
	}{
		"empty path": {
			donor:    &donordata.Provided{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			donor:    &donordata.Provided{},
			url:      "/whatever",
			expected: true,
		},
		"check your lpa when unsure if can sign": {
			donor: &donordata.Provided{
				Type: lpadata.LpaTypePersonalWelfare,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					LifeSustainingTreatment:    task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					AddCorrespondent:           task.StateCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: false,
		},
		"check your lpa when can sign": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{CanSign: form.Yes},
				Type:  lpadata.LpaTypePersonalWelfare,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					LifeSustainingTreatment:    task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					AddCorrespondent:           task.StateCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: true,
		},
		"about payment without task": {
			donor:    &donordata.Provided{LpaID: "123"},
			url:      Paths.AboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePropertyAndAffairs,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					WhenCanTheLpaBeUsed:        task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					CheckYourLpa:               task.StateCompleted,
				},
			},
			url:      Paths.AboutPayment.Format("123"),
			expected: true,
		},
		"identity without task": {
			donor:    &donordata.Provided{},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: false,
		},
		"identity with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePersonalWelfare,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					LifeSustainingTreatment:    task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					CheckYourLpa:               task.StateCompleted,
					PayForLpa:                  task.PaymentStateCompleted,
				},
			},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: true,
		},
		"read lpa without task": {
			donor:    &donordata.Provided{},
			url:      Paths.ReadYourLpa.Format("123"),
			expected: false,
		},
		"read lpa with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Type:                  lpadata.LpaTypePersonalWelfare,
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					LifeSustainingTreatment:    task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					CheckYourLpa:               task.StateCompleted,
					PayForLpa:                  task.PaymentStateCompleted,
				},
			},
			url:      Paths.ReadYourLpa.Format("123"),
			expected: true,
		},
		"your name when identity not set": {
			donor:    &donordata.Provided{},
			url:      Paths.YourName.Format("123"),
			expected: true,
		},
		"your name when identity set": {
			donor: &donordata.Provided{
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			url:      Paths.YourName.Format("123"),
			expected: false,
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
		certificateProvider *certificateproviderdata.Provided
		url                 string
		expected            bool
	}{
		"empty path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 "",
			expected:            false,
		},
		"unexpected path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 "/whatever",
			expected:            true,
		},
		"unrestricted path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 Paths.CertificateProvider.ConfirmYourDetails.Format("123"),
			expected:            true,
		},
		"identity without task": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 Paths.CertificateProvider.IdentityWithOneLogin.Format("123"),
			expected:            false,
		},
		"identity with task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      Paths.CertificateProvider.IdentityWithOneLogin.Format("123"),
			expected: true,
		},
		"provide certificate without task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      Paths.CertificateProvider.ProvideCertificate.Format("123"),
			expected: false,
		},
		"provide certificate with task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.StateCompleted,
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
		attorney *attorneydata.Provided
		url      string
		expected bool
	}{
		"empty path": {
			attorney: &attorneydata.Provided{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			attorney: &attorneydata.Provided{},
			url:      "/whatever",
			expected: true,
		},
		"unrestricted path": {
			attorney: &attorneydata.Provided{},
			url:      Paths.Attorney.ConfirmYourDetails.Format("123"),
			expected: true,
		},
		"sign without task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      Paths.Attorney.Sign.Format("123"),
			expected: false,
		},
		"sign with task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			url:      Paths.Attorney.Sign.Format("123"),
			expected: true,
		},
		"would like second signatory not trust corp": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			url:      Paths.Attorney.WouldLikeSecondSignatory.Format("123"),
			expected: false,
		},
		"would like second signatory as trust corp": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
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
