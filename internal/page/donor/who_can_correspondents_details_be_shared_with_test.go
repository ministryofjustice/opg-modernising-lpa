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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhoCanCorrespondentsDetailsBeSharedWith(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whoCanCorrespondentsDetailsBeSharedWithData{
			App:     testAppData,
			Form:    &whoCanCorrespondentsDetailsBeSharedWithForm{},
			Options: donordata.CorrespondentShareValues,
		}).
		Return(nil)

	err := WhoCanCorrespondentsDetailsBeSharedWith(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhoCanCorrespondentsDetailsBeSharedWithFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whoCanCorrespondentsDetailsBeSharedWithData{
			App: testAppData,
			Form: &whoCanCorrespondentsDetailsBeSharedWithForm{
				Share: actor.CorrespondentShareAttorneys,
			},
			Options: donordata.CorrespondentShareValues,
		}).
		Return(nil)

	err := WhoCanCorrespondentsDetailsBeSharedWith(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Correspondent: actor.Correspondent{Share: actor.CorrespondentShareAttorneys},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhoCanCorrespondentsDetailsBeSharedWithWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := WhoCanCorrespondentsDetailsBeSharedWith(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhoCanCorrespondentsDetailsBeSharedWith(t *testing.T) {
	form := url.Values{
		"share": {actor.CorrespondentShareAttorneys.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Correspondent: actor.Correspondent{
				Share: actor.CorrespondentShareAttorneys,
			},
			Tasks: actor.DonorTasks{
				AddCorrespondent: actor.TaskCompleted,
			},
		}).
		Return(nil)

	err := WhoCanCorrespondentsDetailsBeSharedWith(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWhoCanCorrespondentsDetailsBeSharedWithWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"share": {actor.CorrespondentShareAttorneys.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WhoCanCorrespondentsDetailsBeSharedWith(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostWhoCanCorrespondentsDetailsBeSharedWithWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"share": {"what"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whoCanCorrespondentsDetailsBeSharedWithData) bool {
			return assert.Equal(t, validation.With("share", validation.SelectError{Label: "whoCorrespondentDetailsCanBeSharedWith"}), data.Errors)
		})).
		Return(nil)

	err := WhoCanCorrespondentsDetailsBeSharedWith(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhoCanCorrespondentsDetailsBeSharedWithForm(t *testing.T) {
	form := url.Values{
		"share": {actor.CorrespondentShareAttorneys.String(), actor.CorrespondentShareCertificateProvider.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhoCanCorrespondentsDetailsBeSharedWithForm(r)

	assert.Equal(t, actor.CorrespondentShareAttorneys|actor.CorrespondentShareCertificateProvider, result.Share)
	assert.Nil(t, result.Error)
}

func TestWhoCanCorrespondentsDetailsBeSharedWithFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whoCanCorrespondentsDetailsBeSharedWithForm
		errors validation.List
	}{
		"valid": {
			form: &whoCanCorrespondentsDetailsBeSharedWithForm{},
		},
		"invalid": {
			form: &whoCanCorrespondentsDetailsBeSharedWithForm{
				Error: expectedError,
			},
			errors: validation.With("share", validation.SelectError{Label: "whoCorrespondentDetailsCanBeSharedWith"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
