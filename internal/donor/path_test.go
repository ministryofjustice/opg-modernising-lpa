package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

func TestLpaPathString(t *testing.T) {
	assert.Equal(t, "/lpa/{id}/anything", Path("/anything").String())
}

func TestLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/lpa/abc/anything", Path("/anything").Format("abc"))
}

func TestLpaPathRedirect(t *testing.T) {
	testCases := map[string]struct {
		url      string
		donor    *donordata.Provided
		expected string
	}{
		"redirect": {
			url:      "/",
			donor:    &donordata.Provided{LpaID: "lpa-id"},
			expected: PathConfirmYourIdentity.Format("lpa-id"),
		},
		"redirect with from": {
			url:      "/?from=" + PathRestrictions.Format("lpa-id"),
			donor:    &donordata.Provided{LpaID: "lpa-id"},
			expected: PathRestrictions.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := PathConfirmYourIdentity.Redirect(w, r, appcontext.Data{Lang: localize.En}, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
		})
	}
}

func TestLpaPathRedirectQuery(t *testing.T) {
	testCases := map[string]struct {
		url      string
		donor    *donordata.Provided
		expected string
	}{
		"redirect": {
			url:      "/",
			donor:    &donordata.Provided{LpaID: "lpa-id"},
			expected: PathConfirmYourIdentity.Format("lpa-id") + "?q=1",
		},
		"redirect with from": {
			url:      "/?from=" + PathRestrictions.Format("lpa-id"),
			donor:    &donordata.Provided{LpaID: "lpa-id"},
			expected: PathRestrictions.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := PathConfirmYourIdentity.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, tc.donor, url.Values{"q": {"1"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
		})
	}
}

func TestPathCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		donor    *donordata.Provided
		path     Path
		expected bool
	}{
		"unexpected path": {
			donor:    &donordata.Provided{},
			path:     "/whatever",
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
			path:     PathCheckYourLpa,
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
			path:     PathCheckYourLpa,
			expected: true,
		},
		"about payment without task": {
			donor:    &donordata.Provided{LpaID: "123"},
			path:     PathAboutPayment,
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
			path:     PathAboutPayment,
			expected: true,
		},
		"identity without task": {
			donor:    &donordata.Provided{},
			path:     PathIdentityWithOneLogin,
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
			path:     PathIdentityWithOneLogin,
			expected: true,
		},
		"read lpa without task": {
			donor:    &donordata.Provided{},
			path:     PathReadYourLpa,
			expected: false,
		},
		"read lpa with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Type:             lpadata.LpaTypePersonalWelfare,
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
			path:     PathReadYourLpa,
			expected: true,
		},
		"your name when can change personal details": {
			donor:    &donordata.Provided{},
			path:     PathYourName,
			expected: true,
		},
		"your name when cannot change personal details": {
			donor: &donordata.Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			path:     PathYourName,
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.path.CanGoTo(tc.donor))
		})
	}
}
