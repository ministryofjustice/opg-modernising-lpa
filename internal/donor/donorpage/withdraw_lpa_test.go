package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWithdrawLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &withdrawLpaData{
			App:   testAppData,
			Donor: &donordata.Provided{},
		}).
		Return(nil)

	err := WithdrawLpa(template.Execute, nil, nil, nil, nil, nil, nil, "", "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWithdrawLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := WithdrawLpa(template.Execute, nil, nil, nil, nil, nil, nil, "", "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWithdrawLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedDonor := &donordata.Provided{
		LpaUID: "lpa-uid",
		Donor:  donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:   lpadata.LpaTypePersonalWelfare,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
		WithdrawnAt: testNow,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorWithdrawLPA(r.Context(), "lpa-uid").
		Return(nil)

	err := WithdrawLpa(nil, donorStore, testNowFn, lpaStoreClient, nil, nil, nil, "app://", "http://example.com/certificate-provider")(testAppData, w, r, &donordata.Provided{
		LpaUID: "lpa-uid",
		Donor:  donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:   lpadata.LpaTypePersonalWelfare,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathLpaWithdrawn.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostWithdrawLpaWhenCertificateProviderInvited(t *testing.T) {
	testcases := map[string]struct {
		certificateProvider *certificateproviderdata.Provided
		lpa                 *lpadata.Lpa
	}{
		"certificate provider exists": {
			certificateProvider: &certificateproviderdata.Provided{ContactLanguagePreference: localize.Cy},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "C", LastName: "D",
					Email:                     "a@b.com",
					ContactLanguagePreference: localize.Cy,
				},
			},
		},
		"cannot get certificate provider or does not exist": {
			certificateProvider: &certificateproviderdata.Provided{},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "C", LastName: "D",
					Email: "a@b.com",
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			updatedDonor := &donordata.Provided{
				LpaUID:                       "lpa-uid",
				Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
				Type:                         lpadata.LpaTypePersonalWelfare,
				CertificateProviderInvitedAt: testNow,
				WithdrawnAt:                  testNow,
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), updatedDonor).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendDonorWithdrawLPA(r.Context(), "lpa-uid").
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(mock.Anything).
				Return(tc.lpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(r.Context()).
				Return(tc.certificateProvider, nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				Possessive("A B").
				Return("possessive")
			localizer.EXPECT().
				T(lpadata.LpaTypePersonalWelfare.String()).
				Return("Type")
			localizer.EXPECT().
				FormatDate(testNow).
				Return("formatted date")

			testAppData.Localizer = localizer

			expectedEmail := notify.InformCertificateProviderLPAHasBeenRevoked{
				DonorFullName:                   "A B",
				DonorFullNamePossessive:         "possessive",
				LpaType:                         "type",
				CertificateProviderFullName:     "C D",
				InvitedDate:                     "formatted date",
				CertificateProviderStartPageURL: "http://example.com/certificate-provider",
			}

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(tc.certificateProvider, tc.lpa), "lpa-uid", expectedEmail).
				Return(nil)

			err := WithdrawLpa(nil, donorStore, testNowFn, lpaStoreClient, notifyClient, lpaStoreResolvingService, certificateProviderStore, "app://", "http://example.com/certificate-provider")(testAppData, w, r, &donordata.Provided{
				Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
				LpaUID:                       "lpa-uid",
				Type:                         lpadata.LpaTypePersonalWelfare,
				CertificateProviderInvitedAt: testNow,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.PathLpaWithdrawn.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
		})
	}

}

func TestPostWithdrawLpaWhenVoucherInvited(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaUID:           "lpa-uid",
		Type:             lpadata.LpaTypePropertyAndAffairs,
		Donor:            donordata.Donor{FirstNames: "A", LastName: "B"},
		Voucher:          donordata.Voucher{FirstNames: "C", LastName: "D"},
		VoucherInvitedAt: testNow,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaUID:           "lpa-uid",
			Type:             lpadata.LpaTypePropertyAndAffairs,
			Donor:            donordata.Donor{FirstNames: "A", LastName: "B"},
			Voucher:          donordata.Voucher{FirstNames: "C", LastName: "D"},
			VoucherInvitedAt: testNow,
			WithdrawnAt:      testNow,
		}).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorWithdrawLPA(r.Context(), "lpa-uid").
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToVoucher(provided.Voucher), "lpa-uid", notify.VoucherLpaRevoked{
			DonorFullName:           "A B",
			DonorFullNamePossessive: "A B's",
			InvitedDate:             "2 January 2020",
			LpaType:                 "property and affairs",
			VoucherFullName:         "C D",
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().Possessive("A B").Return("A B's")
	localizer.EXPECT().FormatDate(testNow).Return("2 January 2020")
	localizer.EXPECT().T("property-and-affairs").Return("Property and affairs")

	appData := testAppData
	appData.Localizer = localizer

	err := WithdrawLpa(nil, donorStore, testNowFn, lpaStoreClient, notifyClient, nil, nil, "", "")(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathLpaWithdrawn.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostWithdrawLpaWhenAttorneysInvited(t *testing.T) {
	testcases := map[string]struct {
		lpa               *lpadata.Lpa
		setupNotifyClient func(*lpadata.Lpa, *mockNotifyClient)
	}{
		"no trust corporations": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{FirstNames: "C", LastName: "D", Email: "a@example.com"}},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{FirstNames: "E", LastName: "F", Email: "r@example.com"}},
				},
			},
			setupNotifyClient: func(lpa *lpadata.Lpa, notifyClient *mockNotifyClient) {
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "C D",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "E F",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
			},
		},
		"trust corporation": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				Attorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "C", LastName: "D", Email: "a@example.com"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "t", Email: "t@example.com"},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{FirstNames: "E", LastName: "F", Email: "r@example.com"}},
				},
			},
			setupNotifyClient: func(lpa *lpadata.Lpa, notifyClient *mockNotifyClient) {
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "C D",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "E F",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaTrustCorporation(lpa.Attorneys.TrustCorporation), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "t",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
			},
		},
		"replacement trust corporation": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{FirstNames: "C", LastName: "D", Email: "a@example.com"}},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "E", LastName: "F", Email: "r@example.com"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "t", Email: "t@example.com"},
				},
			},
			setupNotifyClient: func(lpa *lpadata.Lpa, notifyClient *mockNotifyClient) {
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "C D",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "E F",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(mock.Anything, notify.ToLpaTrustCorporation(lpa.ReplacementAttorneys.TrustCorporation), "lpa-uid", notify.AttorneyLpaRevoked{
						DonorFullName:           "A B",
						DonorFullNamePossessive: "A B's",
						InvitedDate:             "2 January 2020",
						LpaType:                 "property and affairs",
						AttorneyFullName:        "t",
						AttorneyStartPageURL:    "http://example.com/attorney-start",
					}).
					Return(nil).
					Once()
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			provided := &donordata.Provided{
				LpaUID:             "lpa-uid",
				AttorneysInvitedAt: testNow,
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaUID:             "lpa-uid",
					AttorneysInvitedAt: testNow,
					WithdrawnAt:        testNow,
				}).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendDonorWithdrawLPA(r.Context(), "lpa-uid").
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			notifyClient := newMockNotifyClient(t)
			tc.setupNotifyClient(tc.lpa, notifyClient)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().Possessive("A B").Return("A B's")
			localizer.EXPECT().FormatDate(testNow).Return("2 January 2020")
			localizer.EXPECT().T("property-and-affairs").Return("Property and affairs")

			appData := testAppData
			appData.Localizer = localizer

			err := WithdrawLpa(nil, donorStore, testNowFn, lpaStoreClient, notifyClient, lpaStoreResolvingService, nil, "http://example.com", "")(appData, w, r, provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.PathLpaWithdrawn.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
		})
	}
}

func TestPostWithdrawLpaWhenNotifyErrors(t *testing.T) {
	testcases := map[string]struct {
		provided            *donordata.Provided
		lpa                 *lpadata.Lpa
		certificateProvider *certificateproviderdata.Provided
	}{
		"voucher": {
			provided: &donordata.Provided{
				LpaUID:           "lpa-uid",
				Type:             lpadata.LpaTypePropertyAndAffairs,
				Donor:            donordata.Donor{FirstNames: "A", LastName: "B"},
				Voucher:          donordata.Voucher{FirstNames: "C", LastName: "D"},
				VoucherInvitedAt: testNow,
			},
		},
		"attorney": {
			provided: &donordata.Provided{
				LpaUID:             "lpa-uid",
				AttorneysInvitedAt: testNow,
			},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{FirstNames: "C", LastName: "D", Email: "a@example.com"}},
				},
			},
		},
		"trust corporation": {
			provided: &donordata.Provided{
				LpaUID:             "lpa-uid",
				AttorneysInvitedAt: testNow,
			},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				Attorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{Name: "t", Email: "t@example.com"},
				},
			},
		},
		"certificate provider": {
			provided: &donordata.Provided{CertificateProviderInvitedAt: testNow},
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePropertyAndAffairs,
				Donor:  lpadata.Donor{FirstNames: "A", LastName: "B"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "C", LastName: "D",
				},
			},
			certificateProvider: &certificateproviderdata.Provided{},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			if tc.lpa != nil {
				lpaStoreResolvingService.EXPECT().
					Get(mock.Anything).
					Return(tc.lpa, nil)
			}

			certificateProviderStore := newMockCertificateProviderStore(t)
			if tc.certificateProvider != nil {
				certificateProviderStore.EXPECT().
					GetAny(mock.Anything).
					Return(tc.certificateProvider, nil)
			}

			localizer := newMockLocalizer(t)
			localizer.EXPECT().Possessive(mock.Anything).Return("A B's")
			localizer.EXPECT().FormatDate(mock.Anything).Return("2 January 2020")
			localizer.EXPECT().T(mock.Anything).Return("Property and affairs")

			testAppData.Localizer = localizer

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)

			err := WithdrawLpa(nil, nil, testNowFn, nil, notifyClient, lpaStoreResolvingService, certificateProviderStore, "", "")(testAppData, w, r, tc.provided)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostWithdrawLpaWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"attorneys":            {AttorneysInvitedAt: testNow},
		"certificate provider": {CertificateProviderInvitedAt: testNow},
	}

	for name, provided := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(mock.Anything).
				Return(nil, expectedError)

			err := WithdrawLpa(nil, nil, testNowFn, nil, nil, lpaStoreResolvingService, nil, "", "")(testAppData, w, r, provided)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostWithdrawLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WithdrawLpa(nil, donorStore, time.Now, nil, nil, nil, nil, "", "")(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostWithdrawLpaWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorWithdrawLPA(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WithdrawLpa(nil, donorStore, time.Now, lpaStoreClient, nil, nil, nil, "", "")(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
