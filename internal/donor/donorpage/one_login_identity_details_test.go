package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOneLoginIdentityDetails(t *testing.T) {
	dob := date.New("1", "2", "3")

	testcases := map[string]struct {
		donorProvided          *actor.DonorProvidedDetails
		expectedDetailsMatch   bool
		expectedDetailsUpdated bool
		url                    string
	}{
		"details match": {
			donorProvided: &actor.DonorProvidedDetails{
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dob, Address: testAddress},
				DonorIdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", DateOfBirth: dob, CurrentAddress: testAddress},
			},
			expectedDetailsMatch: true,
			url:                  "/",
		},
		"details do not match": {
			donorProvided: &actor.DonorProvidedDetails{
				Donor:                 actor.Donor{FirstNames: "a"},
				DonorIdentityUserData: identity.UserData{FirstNames: "b"},
			},
			url: "/",
		},
		"details updated": {
			donorProvided: &actor.DonorProvidedDetails{
				Donor:                 actor.Donor{FirstNames: "a"},
				DonorIdentityUserData: identity.UserData{FirstNames: "b"},
			},
			url:                    "/?detailsUpdated=1",
			expectedDetailsUpdated: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tc.url, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &oneLoginIdentityDetailsData{
					App:            testAppData,
					Form:           form.NewYesNoForm(form.YesNoUnknown),
					DonorProvided:  tc.donorProvided,
					DetailsUpdated: tc.expectedDetailsUpdated,
					DetailsMatch:   tc.expectedDetailsMatch,
				}).
				Return(nil)

			err := OneLoginIdentityDetails(template.Execute, nil)(testAppData, w, r, tc.donorProvided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostOneLoginIdentityDetailsWhenYes(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {form.Yes.String()}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	existingDob := date.New("1", "2", "3")
	identityDob := date.New("4", "5", "6")

	updated := &actor.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 actor.Donor{FirstNames: "b", LastName: "b", DateOfBirth: identityDob, Address: place.Address{Line1: "a"}},
		DonorIdentityUserData: identity.UserData{FirstNames: "b", LastName: "b", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
	}
	updated.UpdateCheckedHash()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updated).
		Return(nil)

	err := OneLoginIdentityDetails(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                 "lpa-id",
		Donor:                 actor.Donor{FirstNames: "a", LastName: "a", DateOfBirth: existingDob, Address: testAddress},
		DonorIdentityUserData: identity.UserData{FirstNames: "b", LastName: "b", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.OneLoginIdentityDetails.Format("lpa-id")+"?detailsUpdated=1", resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenNo(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {form.No.String()}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err := OneLoginIdentityDetails(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WithdrawThisLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenIdentityAndLPADetailsAlreadyMatch(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	err := OneLoginIdentityDetails(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ReadYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenDonorStoreError(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {form.Yes.String()}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := OneLoginIdentityDetails(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostOneLoginIdentityDetailsWhenValidationError(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {""}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesIfWouldLikeToUpdateDetails"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *oneLoginIdentityDetailsData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := OneLoginIdentityDetails(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Donor: actor.Donor{FirstNames: "a"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
