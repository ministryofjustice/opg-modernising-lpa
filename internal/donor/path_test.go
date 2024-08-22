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
			url: "/",
			donor: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePersonalWelfare,
				Tasks: task.DonorTasks{
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
			expected: PathHowToConfirmYourIdentityAndSign.Format("lpa-id"),
		},
		"redirect with from": {
			url:      "/?from=" + PathRestrictions.Format("lpa-id"),
			donor:    &donordata.Provided{LpaID: "lpa-id", Tasks: task.DonorTasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted}},
			expected: PathRestrictions.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			err := PathHowToConfirmYourIdentityAndSign.Redirect(w, r, appcontext.Data{Lang: localize.En}, tc.donor)
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

	err := PathTaskList.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, &donordata.Provided{LpaID: "lpa-id"}, url.Values{"q": {"1"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, PathTaskList.Format("lpa-id")+"?q=1", resp.Header.Get("Location"))
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
				Tasks: task.DonorTasks{
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
			url:      PathCheckYourLpa.Format("123"),
			expected: false,
		},
		"check your lpa when can sign": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{CanSign: form.Yes},
				Type:  lpadata.LpaTypePersonalWelfare,
				Tasks: task.DonorTasks{
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
			url:      PathCheckYourLpa.Format("123"),
			expected: true,
		},
		"about payment without task": {
			donor:    &donordata.Provided{LpaID: "123"},
			url:      PathAboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePropertyAndAffairs,
				Tasks: task.DonorTasks{
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
			url:      PathAboutPayment.Format("123"),
			expected: true,
		},
		"identity without task": {
			donor:    &donordata.Provided{},
			url:      PathIdentityWithOneLogin.Format("123"),
			expected: false,
		},
		"identity with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				Type: lpadata.LpaTypePersonalWelfare,
				Tasks: task.DonorTasks{
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
			url:      PathIdentityWithOneLogin.Format("123"),
			expected: true,
		},
		"read lpa without task": {
			donor:    &donordata.Provided{},
			url:      PathReadYourLpa.Format("123"),
			expected: false,
		},
		"read lpa with tasks": {
			donor: &donordata.Provided{
				Donor: donordata.Donor{
					CanSign: form.Yes,
				},
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Type:                  lpadata.LpaTypePersonalWelfare,
				Tasks: task.DonorTasks{
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
			url:      PathReadYourLpa.Format("123"),
			expected: true,
		},
		"your name when identity not set": {
			donor:    &donordata.Provided{},
			url:      PathYourName.Format("123"),
			expected: true,
		},
		"your name when identity set": {
			donor: &donordata.Provided{
				DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			url:      PathYourName.Format("123"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CanGoTo(tc.donor, tc.url))
		})
	}
}
