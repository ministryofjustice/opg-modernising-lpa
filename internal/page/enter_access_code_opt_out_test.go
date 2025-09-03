package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterAccessCodeOptOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterAccessCodeData{
		App:  testAppData,
		Form: form.NewAccessCodeForm(),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	err := EnterAccessCodeOptOut(template.Execute, newMockAccessCodeStore(t), nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterAccessCodeOptOutOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCodeOptOut(template.Execute, newMockAccessCodeStore(t), nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOptOut(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd 123-4"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234", "Smith")).
		Return(accesscodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa-id"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")),
			ActorUID:    uid,
		}, nil)

	sessionStore := newMockSetLpaDataSessionStore(t)
	sessionStore.EXPECT().
		SetLpaData(r, w, &sesh.LpaDataSession{LpaID: "lpa-id"}).
		Return(nil)

	err := EnterAccessCodeOptOut(nil, accessCodeStore, sessionStore, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, PathDashboard.Format()+"?code=a307b26acffc16da146c9bad4344d510eec887be5e0b78ae6b6f3401730761c3", resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeOptOutErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd 123-4"},
		form.FieldNames.DonorLastName: {"Smith"},
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	testcases := map[string]struct {
		accessCodeStore func() *mockAccessCodeStore
		sessionStore    func() *mockSetLpaDataSessionStore
	}{
		"when accessCodeStore error": {
			accessCodeStore: func() *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accesscodedata.Link{}, expectedError)

				return accessCodeStore
			},
			sessionStore: func() *mockSetLpaDataSessionStore { return nil },
		},
		"when sessionStore error": {
			accessCodeStore: func() *mockAccessCodeStore {
				accessCodeStore := newMockAccessCodeStore(t)
				accessCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

				return accessCodeStore
			},
			sessionStore: func() *mockSetLpaDataSessionStore {
				sessionStore := newMockSetLpaDataSessionStore(t)
				sessionStore.EXPECT().
					SetLpaData(mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return sessionStore
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := EnterAccessCodeOptOut(nil, tc.accessCodeStore(), tc.sessionStore(), actor.TypeAttorney, PathDashboard)(testAppData, w, r)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostEnterAccessCodeOptOutOnAccessCodeStoreNotFoundError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd 1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234", "Smith")).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id"))}, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"}).
				With(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCodeOptOut(template.Execute, accessCodeStore, nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOptOutOnAccessCodeStoreTooManyRequestsError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd 1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234", "Smith")).
		Return(accesscodedata.Link{}, dynamo.ErrTooManyRequests)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.CustomError{Label: "tooManyAccessCodeAttempts"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCodeOptOut(template.Execute, accessCodeStore, nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
