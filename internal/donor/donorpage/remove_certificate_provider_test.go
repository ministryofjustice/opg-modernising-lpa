package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveCertificateProvider(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removeCertificateProviderData{
			App:  testAppData,
			Name: "John Smith",
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemoveCertificateProvider(template.Execute, nil, nil, "")(testAppData, w, r, &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostRemoveCertificateProvider(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
		CertificateProvider: donordata.CertificateProvider{
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
		},
	}

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Delete(r.Context(), provided).
		Return(nil)

	err := RemoveCertificateProvider(nil, service, nil, "")(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseCertificateProvider.FormatQuery("lpa-id", url.Values{
		"removed": {"John Smith"},
	}), resp.Header.Get("Location"))
}

func TestPostRemoveCertificateProviderWhenHasBeenInvited(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
		Type:   lpadata.LpaTypePropertyAndAffairs,
		CertificateProvider: donordata.CertificateProvider{
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
		},
		Donor: donordata.Donor{
			FirstNames: "Darren",
			LastName:   "Donor",
		},
		CertificateProviderInvitedAt: testNow,
	}

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Delete(r.Context(), provided).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), "lpa-uid", notify.CertificateProviderRemoved{
			DonorFullName:                  "Darren Donor",
			CertificateProviderFullName:    "John Smith",
			CertificateProviderInvitedDate: "some date",
			LpaType:                        "some type",
			LpaUID:                         "lpa-uid",
			CertificateProviderStartURL:    "app://" + page.PathCertificateProviderStart.Format(),
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(lpadata.LpaTypePropertyAndAffairs.String()).
		Return("some type")
	localizer.EXPECT().
		FormatDate(testNow).
		Return("some date")

	testAppData := testAppData
	testAppData.Localizer = localizer

	err := RemoveCertificateProvider(nil, service, notifyClient, "app://")(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseCertificateProvider.FormatQuery("lpa-id", url.Values{
		"removed": {"John Smith"},
	}), resp.Header.Get("Location"))
}

func TestPostRemoveCertificateProviderWhenNotifyErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:                        "lpa-id",
		LpaUID:                       "lpa-uid",
		CertificateProviderInvitedAt: testNow,
	}

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("some type")
	localizer.EXPECT().
		FormatDate(mock.Anything).
		Return("some date")

	testAppData := testAppData
	testAppData.Localizer = localizer

	err := RemoveCertificateProvider(nil, nil, notifyClient, "app://")(testAppData, w, r, provided)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostRemoveCertificateProviderWhenServiceErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCertificateProviderService(t)
	service.EXPECT().
		Delete(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCertificateProvider(nil, service, nil, "")(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostRemoveCertificateProviderWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemoveCertificateProvider(nil, nil, nil, "")(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestRemoveCertificateProviderFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveCertificateProvider"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeCertificateProviderData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveCertificateProvider(template.Execute, nil, nil, "")(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
