package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	time "time"

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
		"your name when cannot change personal details because vouch confirmed donor details": {
			donor: &donordata.Provided{
				IdentityUserData:         identity.UserData{Status: identity.StatusInsufficientEvidence},
				DetailsVerifiedByVoucher: true,
			},
			path:     PathYourName,
			expected: false,
		},
		"your date of birth when cannot change personal details because vouch confirmed donor details": {
			donor: &donordata.Provided{
				IdentityUserData:         identity.UserData{Status: identity.StatusInsufficientEvidence},
				DetailsVerifiedByVoucher: true,
			},
			path:     PathYourDateOfBirth,
			expected: false,
		},
		"view lpa": {
			donor:    &donordata.Provided{},
			path:     PathViewLPA,
			expected: false,
		},
		"signed can go to basic pages": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
			},
			path:     PathProgress,
			expected: true,
		},
		"signed open task can go to task list": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
			},
			path:     PathTaskList,
			expected: true,
		},
		"signed completed all tasks can not go to task list": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Type:     lpadata.LpaTypePropertyAndAffairs,
				Donor:    donordata.Donor{CanSign: form.Yes},
				Tasks: donordata.Tasks{
					YourDetails:                task.StateCompleted,
					ChooseAttorneys:            task.StateCompleted,
					ChooseReplacementAttorneys: task.StateCompleted,
					WhenCanTheLpaBeUsed:        task.StateCompleted,
					Restrictions:               task.StateCompleted,
					CertificateProvider:        task.StateCompleted,
					PeopleToNotify:             task.StateCompleted,
					AddCorrespondent:           task.StateCompleted,
					CheckYourLpa:               task.StateCompleted,
					PayForLpa:                  task.PaymentStateCompleted,
					ConfirmYourIdentity:        task.IdentityStateCompleted,
					SignTheLpa:                 task.StateCompleted,
				},
			},
			path:     PathTaskList,
			expected: false,
		},
		"signed payment pending can go to payment page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStatePending,
				},
			},
			path:     PathPayFee,
			expected: true,
		},
		"signed payment complete can not go to payment page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					PayForLpa: task.PaymentStateCompleted,
				},
			},
			path:     PathPayFee,
			expected: false,
		},
		"signed identity pending can go to identity page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStatePending,
				},
			},
			path:     PathIdentityWithOneLogin,
			expected: true,
		},
		"signed identity complete can not go to identity page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
			},
			path:     PathIdentityWithOneLogin,
			expected: false,
		},
		"signed task not completed can go to signing page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					SignTheLpa: task.StateInProgress,
				},
			},
			path:     PathWitnessingAsCertificateProvider,
			expected: true,
		},
		"signed task completed can not go to identity page": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
				Tasks: donordata.Tasks{
					SignTheLpa: task.StateCompleted,
				},
			},
			path:     PathWitnessingAsCertificateProvider,
			expected: false,
		},
		"signed blocks other pages": {
			donor: &donordata.Provided{
				SignedAt: time.Now(),
			},
			path:     PathChooseAttorneys,
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.path.CanGoTo(tc.donor))
		})
	}
}
