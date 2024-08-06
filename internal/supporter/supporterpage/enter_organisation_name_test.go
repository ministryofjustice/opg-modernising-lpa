package supporterpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterOrganisationNameData{
			App:  testAppData,
			Form: &organisationNameForm{},
		}).
		Return(nil)

	err := EnterOrganisationName(nil, template.Execute, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterOrganisationNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterOrganisationName(nil, template.Execute, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationName(t *testing.T) {
	form := url.Values{"name": {"My organisation"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	member := &supporterdata.Member{ID: "a"}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(r.Context()).
		Return(member, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), member, "My organisation").
		Return(&supporterdata.Organisation{ID: "org-id", Name: "My organisation"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{
			IDToken: "id-token",
			Sub:     "random",
			Email:   "name@example.com",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{
			IDToken:          "id-token",
			Sub:              "random",
			Email:            "name@example.com",
			OrganisationID:   "org-id",
			OrganisationName: "My organisation",
		}).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "organisation created", slog.String("organisation_id", "org-id"))

	err := EnterOrganisationName(logger, nil, organisationStore, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathOrganisationCreated.Format(), resp.Header.Get("Location"))
}

func TestPostEnterOrganisationNameWhenSessionStoreSaveError(t *testing.T) {
	form := url.Values{"name": {"My organisation"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(r.Context()).
		Return(&supporterdata.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), mock.Anything, mock.Anything).
		Return(&supporterdata.Organisation{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{
			Sub: "random",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := EnterOrganisationName(logger, nil, organisationStore, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationNameWhenSessionStoreGetError(t *testing.T) {
	form := url.Values{"name": {"My organisation"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(r.Context()).
		Return(&supporterdata.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), mock.Anything, mock.Anything).
		Return(&supporterdata.Organisation{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := EnterOrganisationName(logger, nil, organisationStore, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterStart.Format(), resp.Header.Get("Location"))
}

func TestPostEnterOrganisationNameWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *enterOrganisationNameData) bool {
		return assert.Equal(t, validation.With("name", validation.EnterError{Label: "fullOrganisationOrCompanyName"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterOrganisationNameData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterOrganisationName(nil, template.Execute, nil, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationNameWhenMemberStoreErrors(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := EnterOrganisationName(nil, nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationNameWhenOrganisationStoreErrors(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(r.Context()).
		Return(&supporterdata.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterOrganisationName(nil, nil, organisationStore, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEnterOrganisationNameForm(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readOrganisationNameForm(r, "x")

	assert.Equal(t, "My name", result.Name)
}

func TestEnterOrganisationNameFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *organisationNameForm
		errors validation.List
	}{
		"valid": {
			form: &organisationNameForm{
				Name: "My name",
			},
		},
		"missing": {
			form:   &organisationNameForm{Label: "xyz"},
			errors: validation.With("name", validation.EnterError{Label: "xyz"}),
		},
		"too long": {
			form: &organisationNameForm{
				Name:  strings.Repeat("a", 101),
				Label: "xyz",
			},
			errors: validation.With("name", validation.StringTooLongError{Label: "xyz", Length: 100}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
