package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourLpaLanguage(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourLpaLanguageData{
			App:                testAppData,
			Form:               form.NewYesNoForm(form.YesNoUnknown),
			SelectedLanguage:   localize.En,
			UnselectedLanguage: localize.Cy,
		}).
		Return(nil)

	err := YourLpaLanguage(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{LpaLanguagePreference: localize.En},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourLpaLanguageWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourLpaLanguage(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourLpaLanguageWhenContinue(t *testing.T) {
	for _, lang := range []localize.Lang{localize.En, localize.Cy} {
		t.Run(lang.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.Yes.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := YourLpaLanguage(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{LpaLanguagePreference: lang},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.LpaYourLegalRightsAndResponsibilities.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourLpaLanguageWhenSwitch(t *testing.T) {
	testCases := map[string]struct {
		selectedLanguage localize.Lang
		expectedLanguage localize.Lang
	}{
		"cy": {
			selectedLanguage: localize.En,
			expectedLanguage: localize.Cy,
		},
		"en": {
			selectedLanguage: localize.Cy,
			expectedLanguage: localize.En,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.No.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: actor.Donor{LpaLanguagePreference: tc.expectedLanguage},
				}).
				Return(nil)

			err := YourLpaLanguage(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{LpaLanguagePreference: tc.selectedLanguage},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.LpaYourLegalRightsAndResponsibilities.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourLpaLanguageWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourLpaLanguage(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostYourLpaLanguageWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *yourLpaLanguageData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "whatYouWouldLikeToDo"}), data.Errors)
		})).
		Return(nil)

	err := YourLpaLanguage(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
