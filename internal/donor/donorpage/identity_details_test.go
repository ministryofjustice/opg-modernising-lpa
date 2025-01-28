package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOneLoginIdentityDetailsDataDetailsMatch(t *testing.T) {
	assert.True(t, identityDetailsData{
		FirstNamesMatch:  true,
		LastNameMatch:    true,
		DateOfBirthMatch: true,
		AddressMatch:     true,
	}.DetailsMatch())
	assert.False(t, identityDetailsData{LastNameMatch: true, DateOfBirthMatch: true, AddressMatch: true}.DetailsMatch())
	assert.False(t, identityDetailsData{}.DetailsMatch())
}

func TestGetOneLoginIdentityDetails(t *testing.T) {
	dob := date.New("1", "2", "3")

	testcases := map[string]struct {
		donorProvided            *donordata.Provided
		expectedFirstNamesMatch  bool
		expectedLastNameMatch    bool
		expectedDateOfBirthMatch bool
		expectedAddressMatch     bool
		expectedDetailsUpdated   bool
		url                      string
	}{
		"details match": {
			donorProvided: &donordata.Provided{
				Donor:            donordata.Donor{FirstNames: "A", LastName: "b", DateOfBirth: dob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "B", DateOfBirth: dob, CurrentAddress: testAddress},
			},
			expectedFirstNamesMatch:  true,
			expectedLastNameMatch:    true,
			expectedDateOfBirthMatch: true,
			expectedAddressMatch:     true,
			url:                      "/",
		},
		"details do not match": {
			donorProvided: &donordata.Provided{
				Donor:            donordata.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "b"},
			},
			url: "/",
		},
		"details updated": {
			donorProvided: &donordata.Provided{
				Donor:            donordata.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "b"},
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
				Execute(w, &identityDetailsData{
					App:              testAppData,
					Form:             form.NewYesNoForm(form.YesNoUnknown),
					Provided:         tc.donorProvided,
					DetailsUpdated:   tc.expectedDetailsUpdated,
					FirstNamesMatch:  tc.expectedFirstNamesMatch,
					LastNameMatch:    tc.expectedLastNameMatch,
					DateOfBirthMatch: tc.expectedDateOfBirthMatch,
					AddressMatch:     tc.expectedAddressMatch,
				}).
				Return(nil)

			err := IdentityDetails(template.Execute, nil)(testAppData, w, r, tc.donorProvided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostOneLoginIdentityDetails(t *testing.T) {
	existingDob := date.New("1", "2", "3")
	identityDob := date.New("4", "5", "6")

	updated := &donordata.Provided{
		LpaID:            "lpa-id",
		Donor:            donordata.Donor{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, Address: place.Address{Line1: "a"}},
		IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
		Tasks:            donordata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
	}
	updated.UpdateCheckedHash()

	testcases := map[form.YesNo]struct {
		provided *donordata.Provided
		redirect string
	}{
		form.Yes: {
			provided: updated,
			redirect: donor.PathIdentityDetails.Format("lpa-id") + "?detailsUpdated=1",
		},
		form.No: {
			provided: &donordata.Provided{
				LpaID:                          "lpa-id",
				Donor:                          donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData:               identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				Tasks:                          donordata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
				ContinueWithMismatchedIdentity: true,
			},
			redirect: donor.PathIdentityDetails.Format("lpa-id"),
		},
	}

	for yesNo, tc := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{form.FieldNames.YesNo: {yesNo.String()}}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.provided).
				Return(nil)

			err := IdentityDetails(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
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

	err := IdentityDetails(nil, donorStore)(testAppData, w, r, &donordata.Provided{})
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
		Execute(w, mock.MatchedBy(func(data *identityDetailsData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := IdentityDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Donor: donordata.Donor{FirstNames: "a"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
