package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEditOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	organisation := &supporterdata.Organisation{Name: "what"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &editOrganisationNameData{
			App:  testAppData,
			Form: &organisationNameForm{Name: "what"},
		}).
		Return(nil)

	err := EditOrganisationName(template.Execute, nil)(testAppData, w, r, organisation, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEditOrganisationNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EditOrganisationName(template.Execute, nil)(testAppData, w, r, &supporterdata.Organisation{}, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditOrganisationName(t *testing.T) {
	form := url.Values{"name": {"My organisation"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Put(r.Context(), &supporterdata.Organisation{PK: "ORG", Name: "My organisation"}).
		Return(nil)

	err := EditOrganisationName(nil, organisationStore)(testAppData, w, r, &supporterdata.Organisation{PK: "ORG"}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathOrganisationDetails.Format()+"?updated=name", resp.Header.Get("Location"))
}

func TestPostEditOrganisationNameWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *editOrganisationNameData) bool {
		return assert.Equal(t, validation.With("name", validation.EnterError{Label: "yourOrganisationName"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *editOrganisationNameData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EditOrganisationName(template.Execute, nil)(testAppData, w, r, &supporterdata.Organisation{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditOrganisationNameWhenOrganisationStoreErrors(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EditOrganisationName(nil, organisationStore)(testAppData, w, r, &supporterdata.Organisation{}, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
