package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhatYouCanDoNowVouchNotStarted(t *testing.T) {
	testcases := map[int]struct {
		BannerContent         string
		NewVoucherLabel       string
		ProveOwnIdentityLabel string
		VouchStatusContent    string
		CanHaveVoucher        bool
	}{
		0: {
			BannerContent:         "youHaveNotChosenAnyoneToVouchForYou",
			NewVoucherLabel:       "iHaveSomeoneWhoCanVouch",
			ProveOwnIdentityLabel: "iWillReturnToOneLogin",
			CanHaveVoucher:        true,
		},
		1: {
			BannerContent:         "thePersonYouAskedToVouchHasBeenUnableToContinue",
			NewVoucherLabel:       "iHaveSomeoneElseWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			VouchStatusContent:    "tryVouchingAgainContent",
			CanHaveVoucher:        true,
		},
		2: {
			BannerContent:         "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt",
			NewVoucherLabel:       "iHaveSomeoneElseWhoCanVouch",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
		},
	}

	for vouchAttempts, tc := range testcases {
		t.Run(strconv.Itoa(vouchAttempts), func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			voucherStore := newMockVoucherStore(t)
			voucherStore.EXPECT().
				GetAny(r.Context()).
				Return(nil, dynamo.NotFoundError{})

			donorProvided := &donordata.Provided{VouchAttempts: vouchAttempts}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &whatYouCanDoNowData{
					App: testAppData,
					Form: &whatYouCanDoNowForm{
						Options:        donordata.NoVoucherDecisionValues,
						CanHaveVoucher: tc.CanHaveVoucher,
					},
					Donor:                 donorProvided,
					BannerContent:         tc.BannerContent,
					NewVoucherLabel:       tc.NewVoucherLabel,
					ProveOwnIdentityLabel: tc.ProveOwnIdentityLabel,
					VouchStatusContent:    tc.VouchStatusContent,
				}).
				Return(nil)

			err := WhatYouCanDoNow(template.Execute, nil, voucherStore)(testAppData, w, r, donorProvided)

			assert.Nil(t, err)
		})
	}
}

func TestGetWhatYouCanDoNowVouchStarted(t *testing.T) {
	testcases := map[string]struct {
		BannerContent         string
		VouchStatusContent    string
		ProveOwnIdentityLabel string
		CanHaveVoucher        bool
		Voucher               *voucherdata.Provided
		VouchAttempts         int
	}{
		"entered code only": {
			BannerContent:         "voucherHasNotStartedTheProcess",
			VouchStatusContent:    "voucherHasNotStartedTheProcessContent",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			VouchAttempts:         0,
			Voucher:               &voucherdata.Provided{Email: "a@example.com"},
		},
		"verified donor details": {
			BannerContent:         "voucherHasNotCompletedTheProcess",
			VouchStatusContent:    "voucherHasNotCompletedTheProcessContent",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			VouchAttempts:         1,
			Voucher: &voucherdata.Provided{
				Email: "a@example.com",
				Tasks: voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
			},
		},
		"second voucher entered code only": {
			BannerContent:         "voucherHasNotStartedTheProcess",
			VouchStatusContent:    "voucherHasNotStartedTheProcessContent",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        true,
			VouchAttempts:         1,
			Voucher: &voucherdata.Provided{
				Email: "a@example.com",
			},
		},
		"second voucher verified donor details": {
			BannerContent:         "voucherHasNotCompletedTheProcessSuggestContactVoucher",
			ProveOwnIdentityLabel: "iWillGetOrFindID",
			CanHaveVoucher:        false,
			VouchAttempts:         2,
			Voucher: &voucherdata.Provided{
				Email: "a@example.com",
				Tasks: voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			voucherStore := newMockVoucherStore(t)
			voucherStore.EXPECT().
				GetAny(r.Context()).
				Return(tc.Voucher, nil)

			donorProvided := &donordata.Provided{
				VouchAttempts: tc.VouchAttempts,
				Voucher: donordata.Voucher{
					Allowed: true, FirstNames: "a", LastName: "b",
				},
			}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &whatYouCanDoNowData{
					App: testAppData,
					Form: &whatYouCanDoNowForm{
						Options:        donordata.NoVoucherDecisionValues,
						CanHaveVoucher: tc.CanHaveVoucher,
					},
					BannerContent:         tc.BannerContent,
					VouchStatusContent:    tc.VouchStatusContent,
					NewVoucherLabel:       "iHaveSomeoneElseWhoCanVouch",
					ProveOwnIdentityLabel: tc.ProveOwnIdentityLabel,
					Donor:                 donorProvided,
				}).
				Return(nil)

			err := WhatYouCanDoNow(template.Execute, nil, voucherStore)(testAppData, w, r, donorProvided)

			assert.Nil(t, err)
		})
	}
}

func TestGetWhatYouCanDoNowWhenVoucherStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(&voucherdata.Provided{}, expectedError)

	err := WhatYouCanDoNow(nil, nil, voucherStore)(testAppData, w, r, &donordata.Provided{})

	assert.Error(t, err)
}

func TestGetWhatYouCanDoNowWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(&voucherdata.Provided{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNow(template.Execute, nil, voucherStore)(testAppData, w, r, &donordata.Provided{})

	assert.Error(t, err)
}

func TestPostWhatYouCanDoNow(t *testing.T) {
	testcases := map[donordata.NoVoucherDecision]struct {
		expectedPath  string
		expectedDonor *donordata.Provided
	}{
		donordata.ProveOwnIdentity: {
			expectedPath: donor.PathConfirmYourIdentity.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.No,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		donordata.SelectNewVoucher: {
			expectedPath: donor.PathEnterVoucher.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.Yes,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		donordata.WithdrawLPA: {
			expectedPath: donor.PathWithdrawThisLpa.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:            "lpa-id",
				WantVoucher:      form.No,
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
		donordata.ApplyToCOP: {
			expectedPath: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
			expectedDonor: &donordata.Provided{
				LpaID:                            "lpa-id",
				WantVoucher:                      form.No,
				RegisteringWithCourtOfProtection: true,
				IdentityUserData:                 identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
		},
	}

	for noVoucherDecision, tc := range testcases {
		t.Run(noVoucherDecision.String(), func(t *testing.T) {
			f := url.Values{
				"do-next": {noVoucherDecision.String()},
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.expectedDonor).
				Return(nil)

			err := WhatYouCanDoNow(nil, donorStore, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedPath, resp.Header.Get("Location"))
		})
	}
}

func TestPostWhatYouCanDoNowWhenChangingVoucher(t *testing.T) {
	f := url.Values{
		"do-next": {donordata.ApplyToCOP.String()},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:            "lpa-id",
			IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			Voucher:          donordata.Voucher{Allowed: true},
		}).
		Return(nil)

	err := WhatYouCanDoNow(nil, donorStore, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:            "lpa-id",
		IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
		Voucher:          donordata.Voucher{Allowed: true},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathAreYouSureYouNoLongerNeedVoucher.FormatQuery("lpa-id", url.Values{
		"choice": {donordata.ApplyToCOP.String()},
	}), resp.Header.Get("Location"))
}

func TestPostWhatYouCanDoNowWhenDonorStoreError(t *testing.T) {
	f := url.Values{
		"do-next": {donordata.ApplyToCOP.String()},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatYouCanDoNow(nil, donorStore, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhatYouCanDoNowWhenValidationErrors(t *testing.T) {
	f := url.Values{
		"do-next": {"not a valid value"},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		GetAny(mock.Anything).
		Return(&voucherdata.Provided{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whatYouCanDoNowData) bool {
			return assert.Equal(t, validation.With("do-next", validation.SelectError{Label: "whatYouWouldLikeToDo"}), data.Errors)
		})).
		Return(nil)

	err := WhatYouCanDoNow(template.Execute, nil, voucherStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhatYouCanDoNowForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"do-next": {"  withdraw-lpa  "},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhatYouCanDoNowForm(r, &donordata.Provided{})

	assert.Equal(donordata.WithdrawLPA, result.DoNext)
}

func TestWhatYouCanDoNowFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whatYouCanDoNowForm
		errors validation.List
	}{
		"valid": {
			form: &whatYouCanDoNowForm{
				DoNext: donordata.WithdrawLPA,
			},
		},
		"invalid": {
			form: &whatYouCanDoNowForm{},
			errors: validation.
				With("do-next", validation.SelectError{Label: "whatYouWouldLikeToDo"}),
		},
		"not allowed another vouch": {
			form: &whatYouCanDoNowForm{
				DoNext:         donordata.SelectNewVoucher,
				CanHaveVoucher: false,
			},
			errors: validation.
				With("do-next", validation.CustomError{Label: "youCannotAskAnotherPersonToVouchForYou"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
