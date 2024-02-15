package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetConfirmDonorCanInteractOnline(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDonorCanInteractOnlineData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(expectedError)

	err := ConfirmDonorCanInteractOnline(template.Execute, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDonorCanInteractOnlineWhenYes(t *testing.T) {
	form := url.Values{form.FieldNames.YesNo: {form.Yes.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateLPA(r.Context()).
		Return(&actor.DonorProvidedDetails{LpaID: "lpa-id"}, nil)

	err := ConfirmDonorCanInteractOnline(nil, organisationStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmDonorCanInteractOnlineWhenNo(t *testing.T) {
	form := url.Values{form.FieldNames.YesNo: {form.No.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := ConfirmDonorCanInteractOnline(nil, nil)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ContactOPGForPaperForms.Format(), resp.Header.Get("Location"))
}

func TestPostConfirmDonorCanInteractOnlineWhenValidationError(t *testing.T) {
	f := url.Values{form.FieldNames.YesNo: {"what"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *confirmDonorCanInteractOnlineData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "ifYouWouldLikeToContinueMakingAnOnlineLPA"}), data.Errors)
		}).
		Return(expectedError)

	err := ConfirmDonorCanInteractOnline(template.Execute, nil)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDonorCanInteractOnlineWhenOrganisationStoreError(t *testing.T) {
	form := url.Values{form.FieldNames.YesNo: {form.Yes.String()}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateLPA(r.Context()).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := ConfirmDonorCanInteractOnline(nil, organisationStore)(testAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
