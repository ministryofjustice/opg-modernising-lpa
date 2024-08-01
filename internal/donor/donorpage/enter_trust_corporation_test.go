package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterTrustCorporation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterTrustCorporationData{
			App:  testAppData,
			Form: &enterTrustCorporationForm{},
		}).
		Return(nil)

	err := EnterTrustCorporation(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterTrustCorporationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterTrustCorporationData{
			App:  testAppData,
			Form: &enterTrustCorporationForm{},
		}).
		Return(expectedError)

	err := EnterTrustCorporation(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporation(t *testing.T) {
	form := url.Values{
		"name":           {"Co co."},
		"company-number": {"453345"},
		"email":          {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Attorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
				Name:          "Co co.",
				CompanyNumber: "453345",
				Email:         "name@example.com",
			}},
			Tasks: donordata.Tasks{ChooseAttorneys: actor.TaskInProgress},
		}).
		Return(nil)

	err := EnterTrustCorporation(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.EnterTrustCorporationAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationWhenValidationError(t *testing.T) {
	form := url.Values{
		"company-number": {"453345"},
		"email":          {"name@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterTrustCorporationData) bool {
			return assert.Equal(t, validation.With("name", validation.EnterError{Label: "companyName"}), data.Errors)
		})).
		Return(nil)

	err := EnterTrustCorporation(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"name":           {"Inc co."},
		"company-number": {"64365634"},
		"email":          {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterTrustCorporation(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadEnterTrustCorporationForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"name":           {"  Yoyodyne "},
		"company-number": {"23468723"},
		"email":          {"contact@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterTrustCorporationForm(r)

	assert.Equal("Yoyodyne", result.Name)
	assert.Equal("23468723", result.CompanyNumber)
	assert.Equal("contact@example.com", result.Email)
}

func TestEnterTrustCorporationFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterTrustCorporationForm
		errors validation.List
	}{
		"valid": {
			form: &enterTrustCorporationForm{
				Name:          "A",
				CompanyNumber: "B",
				Email:         "a@b.c",
			},
		},
		"missing all": {
			form: &enterTrustCorporationForm{},
			errors: validation.
				With("name", validation.EnterError{Label: "companyName"}).
				With("company-number", validation.EnterError{Label: "companyNumber"}),
		},
		"invalid email": {
			form: &enterTrustCorporationForm{
				Name:          "A",
				CompanyNumber: "B",
				Email:         "person@",
			},
			errors: validation.With("email", validation.EmailError{Label: "companyEmailAddress"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
