package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testUID = actoruid.New()

func TestGetEnterAccessCode(t *testing.T) {
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

	err := EnterAccessCode(template.Execute, nil, nil, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterAccessCodeOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterAccessCodeData{
		App:  testAppData,
		Form: form.NewAccessCodeForm(),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(expectedError)

	err := EnterAccessCode(template.Execute, nil, nil, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCode(t *testing.T) {
	testcases := map[string]struct {
		lpaOwnerKey dynamo.LpaOwnerKeyType
		appSession  *appcontext.Session
	}{
		"donor": {
			lpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")),
			appSession:  &appcontext.Session{SessionID: "aGV5", LpaID: "lpa-id"},
		},
		"organisation": {
			lpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("123")),
			appSession:  &appcontext.Session{SessionID: "aGV5", LpaID: "lpa-id", OrganisationID: "123"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				form.FieldNames.AccessCode:    {"abcd1234"},
				form.FieldNames.DonorLastName: {"Smith"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			session := &sesh.LoginSession{
				Sub:     "hey",
				Email:   "a@example.com",
				HasLPAs: true,
			}

			accessCode := accesscodedata.Link{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: tc.lpaOwnerKey,
				ActorUID:    testUID,
				LpaUID:      "lpa-uid",
			}

			lpa := &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{LastName: "Smith"},
			}

			accessCodeStore := newMockAccessCodeStore(t)
			accessCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234")).
				Return(accessCode, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Sub: "hey", Email: "a@example.com"}, nil)
			sessionStore.EXPECT().
				SetLogin(r, w, session).
				Return(nil)

			newCtx := mock.MatchedBy(func(ctx context.Context) bool {
				session, _ := appcontext.SessionFromContext(ctx)

				return assert.Equal(t, tc.appSession, session)
			})

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(newCtx).
				Return(lpa, nil)

			next := newMockEnterAccessCodeHandler(t)
			next.EXPECT().
				Execute(testAppData, w, mock.Anything, session, lpa, accessCode).
				Return(nil)

			err := EnterAccessCode(nil, accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeAttorney, next.Execute)(testAppData, w, r)
			assert.Nil(t, err)
		})
	}
}

func TestPostEnterAccessCodeWhenDonorLastNameIncorrect(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smithy"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCode := accesscodedata.Link{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")),
		ActorUID:    testUID,
		LpaUID:      "lpa-uid",
	}

	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		Donor:  lpadata.Donor{LastName: "Smith"},
	}

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(accessCode, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey", Email: "a@example.com"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCode(template.Execute, accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeAttorney, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnAccessCodeStoreError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {" abcd1234  "},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234")).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, expectedError)

	err := EnterAccessCode(nil, accessCodeStore, nil, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnAccessCodeStoreNotFoundError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd 1-234 "},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"}), data.Errors)
		})).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234")).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, dynamo.NotFoundError{})

	err := EnterAccessCode(template.Execute, accessCodeStore, nil, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnLpaStoreResolvingServiceError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := EnterAccessCode(nil, accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeAttorney, nil)(testAppData, w, r)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnSessionGetError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, accesscodedata.HashedFromString("abcd1234")).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, expectedError)

	err := EnterAccessCode(nil, accessCodeStore, sessionStore, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnSessionSetError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{
			Donor: lpadata.Donor{
				LastName: "Smith",
			},
		}, nil)

	err := EnterAccessCode(nil, accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeAttorney, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {""},
		form.FieldNames.DonorLastName: {"abc"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.EnterError{Label: "yourAccessCode"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCode(template.Execute, nil, nil, nil, actor.TypeAttorney, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
