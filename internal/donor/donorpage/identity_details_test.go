package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityDetailsDataDetailsMatch(t *testing.T) {
	assert.True(t, identityDetailsData{
		FirstNamesMatch:  true,
		LastNameMatch:    true,
		DateOfBirthMatch: true,
		AddressMatch:     true,
	}.DetailsMatch())
	assert.False(t, identityDetailsData{LastNameMatch: true, DateOfBirthMatch: true, AddressMatch: true}.DetailsMatch())
	assert.False(t, identityDetailsData{}.DetailsMatch())
}

func TestGetIdentityDetails(t *testing.T) {
	dob := date.New("1", "2", "3")

	testcases := map[string]struct {
		donorProvided            *donordata.Provided
		expectedFirstNamesMatch  bool
		expectedLastNameMatch    bool
		expectedDateOfBirthMatch bool
		expectedAddressMatch     bool
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
					FirstNamesMatch:  tc.expectedFirstNamesMatch,
					LastNameMatch:    tc.expectedLastNameMatch,
					DateOfBirthMatch: tc.expectedDateOfBirthMatch,
					AddressMatch:     tc.expectedAddressMatch,
				}).
				Return(nil)

			err := IdentityDetails(template.Execute, nil, nil)(testAppData, w, r, tc.donorProvided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostIdentityDetails(t *testing.T) {
	existingDob := date.New("1", "2", "3")
	identityDob := date.New("4", "5", "6")

	testcases := map[string]struct {
		yesNo       form.YesNo
		provided    *donordata.Provided
		updated     *donordata.Provided
		eventClient func(*testing.T) *mockEventClient
		redirect    string
	}{
		"yes can change": {
			yesNo: form.Yes,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
			},
			updated: &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, Address: place.Address{Line1: "a"}},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				Tasks: donordata.Tasks{
					CheckYourLpa:        task.StateInProgress,
					ConfirmYourIdentity: task.IdentityStateCompleted,
				},
				IdentityDetailsCausedCheck: true,
			},
			eventClient: func(*testing.T) *mockEventClient { return nil },
			redirect:    donor.PathIdentityDetailsUpdated.Format("lpa-id"),
		},
		"no can change": {
			yesNo: form.No,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
			},
			updated: &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				Tasks:            donordata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
			},
			eventClient: func(*testing.T) *mockEventClient { return nil },
			redirect:    donor.PathRegisterWithCourtOfProtection.Format("lpa-id"),
		},
		"yes cannot change": {
			yesNo: form.Yes,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				SignedAt:         time.Now(),
			},
			eventClient: func(*testing.T) *mockEventClient { return nil },
			redirect:    donor.PathWithdrawThisLpa.Format("lpa-id"),
		},
		"no cannot change": {
			yesNo: form.No,
			provided: &donordata.Provided{
				LpaID:            "lpa-id",
				LpaUID:           "lpa-uid",
				Donor:            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData: identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				SignedAt:         testNow,
			},
			updated: &donordata.Provided{
				LpaID:                            "lpa-id",
				LpaUID:                           "lpa-uid",
				Donor:                            donordata.Donor{FirstNames: "b", LastName: "b", DateOfBirth: existingDob, Address: testAddress},
				IdentityUserData:                 identity.UserData{FirstNames: "B", LastName: "B", DateOfBirth: identityDob, CurrentAddress: place.Address{Line1: "a"}},
				Tasks:                            donordata.Tasks{ConfirmYourIdentity: task.IdentityStateProblem},
				RegisteringWithCourtOfProtection: true,
				SignedAt:                         testNow,
			},
			eventClient: func(*testing.T) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendRegisterWithCourtOfProtection(mock.Anything, event.RegisterWithCourtOfProtection{
						UID: "lpa-uid",
					}).
					Return(nil)

				return eventClient
			},
			redirect: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{form.FieldNames.YesNo: {tc.yesNo.String()}}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			donorStore := newMockDonorStore(t)
			if tc.updated != nil {
				donorStore.EXPECT().
					Put(r.Context(), tc.updated).
					Return(nil)
			}

			err := IdentityDetails(nil, donorStore, tc.eventClient(t))(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostIdentityDetailsWhenDonorStoreError(t *testing.T) {
	for _, yesNo := range []form.YesNo{form.Yes, form.No} {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{form.FieldNames.YesNo: {yesNo.String()}}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), mock.Anything).
				Return(expectedError)

			err := IdentityDetails(nil, donorStore, nil)(testAppData, w, r, &donordata.Provided{})
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostIdentityDetailsWhenEventClientError(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {form.No.String()}}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendRegisterWithCourtOfProtection(r.Context(), mock.Anything).
		Return(expectedError)

	err := IdentityDetails(nil, nil, eventClient)(testAppData, w, r, &donordata.Provided{SignedAt: time.Now()})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityDetailsWhenValidationError(t *testing.T) {
	testcases := map[string]struct {
		provided *donordata.Provided
		label    string
	}{
		"can change": {
			provided: &donordata.Provided{Donor: donordata.Donor{FirstNames: "a"}},
			label:    "yesIfWouldLikeToUpdateDetails",
		},
		"cannot change": {
			provided: &donordata.Provided{Donor: donordata.Donor{FirstNames: "a"}, SignedAt: time.Now()},
			label:    "yesToRevokeThisLpaAndMakeNew",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{form.FieldNames.YesNo: {""}}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: tc.label})

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *identityDetailsData) bool {
					return assert.Equal(t, validationError, data.Errors)
				})).
				Return(nil)

			err := IdentityDetails(template.Execute, nil, nil)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
