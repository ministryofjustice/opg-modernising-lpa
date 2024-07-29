package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
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
			Options: donordata.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowDoYouKnowYourCertificateProviderFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		Relationship: actor.Personally,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howDoYouKnowYourCertificateProviderData{
			App:                 testAppData,
			CertificateProvider: certificateProvider,
			Form:                &howDoYouKnowYourCertificateProviderForm{How: actor.Personally},
			Options:             donordata.CertificateProviderRelationshipValues,
		}).
		Return(nil)

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowDoYouKnowYourCertificateProvider(t *testing.T) {
	testCases := map[string]struct {
		form                       url.Values
		certificateProviderDetails actor.CertificateProvider
		redirect                   page.LpaPath
	}{
		"professionally": {
			form: url.Values{"how": {actor.Professionally.String()}},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames:   "John",
				Relationship: actor.Professionally,
			},
			redirect: page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole,
		},
		"personally": {
			form: url.Values{"how": {actor.Personally.String()}},
			certificateProviderDetails: actor.CertificateProvider{
				FirstNames:   "John",
				Relationship: actor.Personally,
			},
			redirect: page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:               "lpa-id",
					CertificateProvider: tc.certificateProviderDetails,
					Tasks: actor.DonorTasks{
						YourDetails:     actor.TaskCompleted,
						ChooseAttorneys: actor.TaskCompleted,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:               "lpa-id",
				CertificateProvider: actor.CertificateProvider{FirstNames: "John"},
				Tasks: actor.DonorTasks{
					YourDetails:     actor.TaskCompleted,
					ChooseAttorneys: actor.TaskCompleted,
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
		existingCertificateProviderDetails actor.CertificateProvider
		updatedCertificateProviderDetails  actor.CertificateProvider
		redirect                           page.LpaPath
		taskState                          actor.TaskState
	}{
		"personally to professionally": {
			form: url.Values{"how": {actor.Professionally.String()}},
			existingCertificateProviderDetails: actor.CertificateProvider{
				RelationshipLength: actor.GreaterThanEqualToTwoYears,
				Relationship:       actor.Personally,
				Address:            testAddress,
			},
			updatedCertificateProviderDetails: actor.CertificateProvider{
				Relationship: actor.Professionally,
				Address:      place.Address{},
			},
			redirect: page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole,
		},
		"professionally to personally": {
			form: url.Values{"how": {actor.Personally.String()}},
			existingCertificateProviderDetails: actor.CertificateProvider{
				Relationship: actor.Professionally,
				Address:      testAddress,
			},
			updatedCertificateProviderDetails: actor.CertificateProvider{
				Relationship: actor.Personally,
				Address:      place.Address{},
			},
			redirect: page.Paths.HowLongHaveYouKnownCertificateProvider,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:               "lpa-id",
					CertificateProvider: tc.updatedCertificateProviderDetails,
					Tasks: actor.DonorTasks{
						YourDetails:         actor.TaskCompleted,
						ChooseAttorneys:     actor.TaskCompleted,
						CertificateProvider: actor.TaskInProgress,
					},
				}).
				Return(nil)

			err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:               "lpa-id",
				CertificateProvider: tc.existingCertificateProviderDetails,
				Tasks: actor.DonorTasks{
					YourDetails:         actor.TaskCompleted,
					ChooseAttorneys:     actor.TaskCompleted,
					CertificateProvider: actor.TaskCompleted,
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
		"how": {actor.Personally.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowDoYouKnowYourCertificateProvider(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

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

	err := HowDoYouKnowYourCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowDoYouKnowYourCertificateProviderForm(t *testing.T) {
	form := url.Values{
		"how": {actor.Personally.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readHowDoYouKnowYourCertificateProviderForm(r)

	assert.Equal(t, actor.Personally, result.How)
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
