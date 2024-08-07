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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowLongHaveYouKnownCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howLongHaveYouKnownCertificateProviderData{
			App:     testAppData,
			Options: donordata.CertificateProviderRelationshipLengthValues,
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowLongHaveYouKnownCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := donordata.CertificateProvider{RelationshipLength: donordata.GreaterThanEqualToTwoYears}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howLongHaveYouKnownCertificateProviderData{
			App:                 testAppData,
			CertificateProvider: certificateProvider,
			RelationshipLength:  donordata.GreaterThanEqualToTwoYears,
			Options:             donordata.CertificateProviderRelationshipLengthValues,
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{CertificateProvider: certificateProvider})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowLongHaveYouKnownCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howLongHaveYouKnownCertificateProviderData{
			App:     testAppData,
			Options: donordata.CertificateProviderRelationshipLengthValues,
		}).
		Return(expectedError)

	err := HowLongHaveYouKnownCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowLongHaveYouKnownCertificateProviderMoreThan2Years(t *testing.T) {
	form := url.Values{
		"relationship-length": {donordata.GreaterThanEqualToTwoYears.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			Attorneys:           donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}}},
			AttorneyDecisions:   donordata.AttorneyDecisions{How: lpadata.Jointly},
			CertificateProvider: donordata.CertificateProvider{RelationshipLength: donordata.GreaterThanEqualToTwoYears},
			Tasks:               donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID:             "lpa-id",
		Attorneys:         donordata.Attorneys{Attorneys: []donordata.Attorney{{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c"}, DateOfBirth: date.New("1990", "1", "1")}}},
		AttorneyDecisions: donordata.AttorneyDecisions{How: lpadata.Jointly},
		Tasks:             donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowLongHaveYouKnownCertificateProviderLessThan2Years(t *testing.T) {
	form := url.Values{
		"relationship-length": {"lt-2-years"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := HowLongHaveYouKnownCertificateProvider(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseNewCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"relationship-length": {donordata.GreaterThanEqualToTwoYears.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowLongHaveYouKnownCertificateProvider(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostHowLongHaveYouKnownCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howLongHaveYouKnownCertificateProviderData{
			App:     testAppData,
			Errors:  validation.With("relationship-length", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
			Options: donordata.CertificateProviderRelationshipLengthValues,
		}).
		Return(nil)

	err := HowLongHaveYouKnownCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowLongHaveYouKnownCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"relationship-length": {donordata.GreaterThanEqualToTwoYears.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowLongHaveYouKnownCertificateProviderForm(r)

	assert.Equal(t, donordata.GreaterThanEqualToTwoYears, result.RelationshipLength)
}

func TestHowLongHaveYouKnownCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howLongHaveYouKnownCertificateProviderForm
		errors validation.List
	}{
		"valid": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				RelationshipLength: donordata.GreaterThanEqualToTwoYears,
			},
		},
		"invalid": {
			form: &howLongHaveYouKnownCertificateProviderForm{
				Error: expectedError,
			},
			errors: validation.With("relationship-length", validation.SelectError{Label: "howLongYouHaveKnownCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
