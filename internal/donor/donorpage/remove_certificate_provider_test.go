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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveCertificateProvider(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := donordata.CertificateProvider{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removeCertificateProviderData{
			App:  testAppData,
			Name: "John Smith",
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemoveCertificateProvider(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{CertificateProvider: certificateProvider})
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

	certificateProvider := donordata.CertificateProvider{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(r.Context(), certificateProvider).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{LpaID: "lpa-id", LpaUID: "lpa-uid"}).
		Return(nil)

	err := RemoveCertificateProvider(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		LpaID:               "lpa-id",
		LpaUID:              "lpa-uid",
		CertificateProvider: certificateProvider,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveCertificateProviderWhenReuseClientErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCertificateProvider(nil, nil, reuseStore)(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostRemoveCertificateProviderWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemoveCertificateProvider(nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveCertificateProviderErrorOnPutStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCertificateProvider(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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

	err := RemoveCertificateProvider(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
