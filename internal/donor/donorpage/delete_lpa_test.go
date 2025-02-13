package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDeleteLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &deleteLpaData{
			App:   testAppData,
			Donor: &donordata.Provided{},
		}).
		Return(nil)

	err := DeleteLpa(template.Execute, nil, nil, "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDeleteLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DeleteLpa(template.Execute, nil, nil, "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDeleteLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Delete(r.Context()).
		Return(nil)

	err := DeleteLpa(nil, donorStore, nil, "app://")(testAppData, w, r, &donordata.Provided{
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
	assert.Equal(t, page.PathLpaDeleted.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostDeleteLpaWhenCertificateProviderInvited(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaUID:                       "lpa-uid",
		Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:                         lpadata.LpaTypePersonalWelfare,
		CertificateProviderInvitedAt: testNow,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Delete(r.Context()).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive("a b").
		Return("possessive")
	localizer.EXPECT().
		T(lpadata.LpaTypePersonalWelfare.String()).
		Return("Type")
	localizer.EXPECT().
		FormatDate(testNow).
		Return("formatted date")

	testAppData.Localizer = localizer

	email := notify.InformCertificateProviderLPAHasBeenDeleted{
		DonorFullName:                   "a b",
		DonorFullNamePossessive:         "possessive",
		LpaType:                         "type",
		CertificateProviderFullName:     "c d",
		InvitedDate:                     "formatted date",
		CertificateProviderStartPageURL: "app://" + "/certificate-provider-start",
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToCertificateProvider(donor.CertificateProvider), "lpa-uid", email).
		Return(nil)

	err := DeleteLpa(nil, donorStore, notifyClient, "app://")(testAppData, w, r, &donordata.Provided{
		LpaUID:                       "lpa-uid",
		Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:                         lpadata.LpaTypePersonalWelfare,
		CertificateProviderInvitedAt: testNow,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathLpaDeleted.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostDeleteLpaWhenVoucherInvited(t *testing.T) {
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
		Delete(r.Context()).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToVoucher(provided.Voucher), "lpa-uid", notify.VoucherLpaDeleted{
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

	err := DeleteLpa(nil, donorStore, notifyClient, "")(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathLpaDeleted.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostDeleteLpaWhenNotifyErrors(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"voucher invited":              {VoucherInvitedAt: testNow},
		"certificate provider invited": {CertificateProviderInvitedAt: testNow},
	}

	for name, donorProvided := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().Possessive(mock.Anything).Return("A B's")
			localizer.EXPECT().FormatDate(mock.Anything).Return("2 January 2020")
			localizer.EXPECT().T(mock.Anything).Return("Property and affairs")

			appData := testAppData
			appData.Localizer = localizer

			err := DeleteLpa(nil, nil, notifyClient, "")(appData, w, r, donorProvided)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostDeleteLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Delete(r.Context()).
		Return(expectedError)

	err := DeleteLpa(nil, donorStore, nil, "")(testAppData, w, r, &donordata.Provided{})

	assert.ErrorContains(t, err, "error deleting lpa: err")
}
