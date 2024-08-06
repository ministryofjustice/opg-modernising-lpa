package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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

func TestGetHowDoYouKnowYourCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howDoYouKnowYourCertificateProviderData{
			App:     testAppData,
			Form:    &howDoYouKnowYourCertificateProviderForm{},
			Options: lpadata.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := donordata.CertificateProvider{
		Relationship: lpadata.Personally,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howDoYouKnowYourCertificateProviderData{
			App:                 testAppData,
			CertificateProvider: certificateProvider,
			Form:                &howDoYouKnowYourCertificateProviderForm{How: lpadata.Personally},
			Options:             lpadata.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		CertificateProvider: certificateProvider,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowDoYouKnowYourCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails donordata.CertificateProvider
		redirect                   donor.Path
	}{
		"professionally": {
			form: url.Values{"how": {lpadata.Professionally.String()}},
			certificateProviderDetails: donordata.CertificateProvider{
				FirstNames:   "John",
				Relationship: lpadata.Professionally,
			},
			redirect: donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole,
		},
		"personally": {
			form: url.Values{"how": {lpadata.Personally.String()}},
			certificateProviderDetails: donordata.CertificateProvider{
				FirstNames:   "John",
				Relationship: lpadata.Personally,
			},
			redirect: donor.PathHowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					CertificateProvider: tc.certificateProviderDetails,
					Tasks: donordata.Tasks{
						YourDetails:     task.StateCompleted,
						ChooseAttorneys: task.StateCompleted,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:               "lpa-id",
				CertificateProvider: donordata.CertificateProvider{FirstNames: "John"},
				Tasks: donordata.Tasks{
					YourDetails:     task.StateCompleted,
					ChooseAttorneys: task.StateCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowDoYouKnowYourCertificateProviderWhenSwitchingRelationship(t *testing.T) {
	testCases := map[string]struct {
		form                               url.Values
		existingCertificateProviderDetails donordata.CertificateProvider
		updatedCertificateProviderDetails  donordata.CertificateProvider
		redirect                           donor.Path
		taskState                          task.State
	}{
		"personally to professionally": {
			form: url.Values{"how": {lpadata.Professionally.String()}},
			existingCertificateProviderDetails: donordata.CertificateProvider{
				RelationshipLength: donordata.GreaterThanEqualToTwoYears,
				Relationship:       lpadata.Personally,
				Address:            testAddress,
			},
			updatedCertificateProviderDetails: donordata.CertificateProvider{
				Relationship: lpadata.Professionally,
				Address:      place.Address{},
			},
			redirect: donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole,
		},
		"professionally to personally": {
			form: url.Values{"how": {lpadata.Personally.String()}},
			existingCertificateProviderDetails: donordata.CertificateProvider{
				Relationship: lpadata.Professionally,
				Address:      testAddress,
			},
			updatedCertificateProviderDetails: donordata.CertificateProvider{
				Relationship: lpadata.Personally,
				Address:      place.Address{},
			},
			redirect: donor.PathHowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					CertificateProvider: tc.updatedCertificateProviderDetails,
					Tasks: donordata.Tasks{
						YourDetails:         task.StateCompleted,
						ChooseAttorneys:     task.StateCompleted,
						CertificateProvider: task.StateInProgress,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:               "lpa-id",
				CertificateProvider: tc.existingCertificateProviderDetails,
				Tasks: donordata.Tasks{
					YourDetails:         task.StateCompleted,
					ChooseAttorneys:     task.StateCompleted,
					CertificateProvider: task.StateCompleted,
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowDoYouKnowYourCertificateProviderWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"how": {lpadata.Personally.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostHowDoYouKnowYourCertificateProviderWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howDoYouKnowYourCertificateProviderData) bool {
			return assert.Equal(t, validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}), data.Errors)
		})).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowDoYouKnowYourCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how": {lpadata.Personally.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowDoYouKnowYourCertificateProviderForm(r)

	assert.Equal(t, lpadata.Personally, result.How)
}

func TestHowDoYouKnowYourCertificateProviderFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howDoYouKnowYourCertificateProviderForm
		errors validation.List
	}{
		"valid": {
			form: &howDoYouKnowYourCertificateProviderForm{},
		},
		"invalid": {
			form: &howDoYouKnowYourCertificateProviderForm{
				Error: expectedError,
			},
			errors: validation.With("how", validation.SelectError{Label: "howYouKnowCertificateProvider"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
