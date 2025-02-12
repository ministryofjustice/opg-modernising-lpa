package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

	err := WithdrawLpa(template.Execute, nil, nil, nil, nil, "")(testAppData, w, r, &donordata.Provided{})
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

	err := WithdrawLpa(template.Execute, nil, nil, nil, nil, "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWithdrawLpa(t *testing.T) {
	now := time.Now()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedDonor := &donordata.Provided{
		LpaUID:                       "lpa-uid",
		Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:                         lpadata.LpaTypePersonalWelfare,
		CertificateProviderInvitedAt: testNow,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
		WithdrawnAt: now,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorWithdrawLPA(r.Context(), "lpa-uid").
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

	expectedEmail := notify.InformCertificateProviderLPAHasBeenRevoked{
		DonorFullName:                   "a b",
		DonorFullNamePossessive:         "possessive",
		LpaType:                         "type",
		CertificateProviderFullName:     "c d",
		InvitedDate:                     "formatted date",
		CertificateProviderStartPageURL: "app://" + "/certificate-provider-start",
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToCertificateProvider(donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		}), "lpa-uid", expectedEmail).
		Return(nil)

	err := WithdrawLpa(nil, donorStore, func() time.Time { return now }, lpaStoreClient, notifyClient, "app://")(testAppData, w, r, &donordata.Provided{
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
	assert.Equal(t, page.PathLpaWithdrawn.Format()+"?uid=lpa-uid", resp.Header.Get("Location"))
}

func TestPostWithdrawLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WithdrawLpa(nil, donorStore, time.Now, nil, nil, "")(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
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

	err := WithdrawLpa(nil, donorStore, time.Now, lpaStoreClient, nil, "")(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestPostWithdrawLpaWhenNotifyClientError(t *testing.T) {
	now := time.Now()

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendDonorWithdrawLPA(mock.Anything, mock.Anything).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Possessive(mock.Anything).
		Return("possessive")
	localizer.EXPECT().
		T(mock.Anything).
		Return("Type")
	localizer.EXPECT().
		FormatDate(mock.Anything).
		Return("formatted date")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := WithdrawLpa(nil, donorStore, func() time.Time { return now }, lpaStoreClient, notifyClient, "app://")(testAppData, w, r, &donordata.Provided{
		LpaUID:                       "lpa-uid",
		Donor:                        donordata.Donor{FirstNames: "a", LastName: "b"},
		Type:                         lpadata.LpaTypePersonalWelfare,
		CertificateProviderInvitedAt: testNow,
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "c", LastName: "d", Email: "a@b.com",
		},
	})

	assert.ErrorContains(t, err, "error sending LPA revoked email to certificate provider: err")
}
