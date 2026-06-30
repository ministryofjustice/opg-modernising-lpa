package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCompanyNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &companyNumberData{
			App:  testTrustCorporationAppData,
			Form: newCompanyNumberForm(testTrustCorporationAppData.Localizer),
		}).
		Return(nil)

	err := CompanyNumber(template.Execute, nil)(testTrustCorporationAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCompanyNumberFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	form := newCompanyNumberForm(testTrustCorporationAppData.Localizer)
	form.CompanyNumber.Set("12345678")

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &companyNumberData{
			App:  testTrustCorporationAppData,
			Form: form,
		}).
		Return(nil)

	err := CompanyNumber(template.Execute, nil)(testTrustCorporationAppData, w, r, &attorneydata.Provided{
		CompanyNumber: "12345678",
	}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCompanyNumberWhenNotTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CompanyNumber(nil, nil)(testAppData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetCompanyNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := CompanyNumber(template.Execute, nil)(testTrustCorporationAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCompanyNumber(t *testing.T) {
	form := url.Values{
		"company-number": {"12345678"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), &attorneydata.Provided{
			LpaID:         "lpa-id",
			CompanyNumber: "12345678",
			Tasks: attorneydata.Tasks{
				ConfirmYourDetails: task.StateInProgress,
			},
		}).
		Return(nil)

	err := CompanyNumber(nil, attorneyStore)(testTrustCorporationAppData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathPhoneNumber.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCompanyNumberWhenNotTrustCorporation(t *testing.T) {
	form := url.Values{
		"company-number": {"12345678"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := CompanyNumber(nil, nil)(testAppData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCompanyNumberWhenValidationError(t *testing.T) {
	form := url.Values{
		"company-number": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *companyNumberData) bool {
		return assert.Equal(t, []forms.Field{data.Form.CompanyNumber.Field}, data.Form.Errors) &&
			assert.Equal(t, "errorEnter:Label=yourCompanyNumber", data.Form.CompanyNumber.Error.Format(testTrustCorporationAppData.Localizer))
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *companyNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := CompanyNumber(template.Execute, nil)(testTrustCorporationAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCompanyNumberWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		"company-number": {"12345678"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CompanyNumber(nil, attorneyStore)(testTrustCorporationAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
