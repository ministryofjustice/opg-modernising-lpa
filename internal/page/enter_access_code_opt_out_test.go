package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
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

	err := EnterAccessCodeOptOut(template.Execute, newMockShareCodeStore(t), nil, nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

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

	err := EnterAccessCodeOptOut(template.Execute, newMockShareCodeStore(t), nil, nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

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

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa-id"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")),
			ActorUID:    uid,
		}, nil)

	sessionStore := newMockSetLpaDataSessionStore(t)
	sessionStore.EXPECT().
		SetLpaData(r, w, &sesh.LpaDataSession{LpaID: "lpa-id"}).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{
			Donor: lpadata.Donor{
				LastName: "Smith",
			},
		}, nil)

	err := EnterAccessCodeOptOut(nil, shareCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeAttorney, PathDashboard)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, PathDashboard.Format()+"?code=e9cee71ab932fde863338d08be4de9dfe39ea049bdafb342ce659ec5450b69ae", resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeOptOutWhenDonorLastNameIncorrect(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd 123-4"},
		form.FieldNames.DonorLastName: {"Smithy"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{
			LpaKey:      dynamo.LpaKey("lpa-id"),
			LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")),
			ActorUID:    uid,
		}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{
			Donor: lpadata.Donor{
				LastName: "Smith",
			},
		}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCodeOptOut(template.Execute, shareCodeStore, nil, lpaStoreResolvingService, actor.TypeAttorney, PathDashboard)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
		shareCodeStore           func() *mockShareCodeStore
		lpaStoreResolvingService func() *mockLpaStoreResolvingService
		sessionStore             func() *mockSetLpaDataSessionStore
	}{
		"when shareCodeStore error": {
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(sharecodedata.Link{}, expectedError)

				return shareCodeStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService { return nil },
			sessionStore:             func() *mockSetLpaDataSessionStore { return nil },
		},
		"when lpaStoreResolvingService error": {
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

				return shareCodeStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(mock.Anything).
					Return(nil, expectedError)

				return lpaStoreResolvingService
			},
			sessionStore: func() *mockSetLpaDataSessionStore { return nil },
		},
		"when sessionStore error": {
			shareCodeStore: func() *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

				return shareCodeStore
			},
			lpaStoreResolvingService: func() *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(mock.Anything).
					Return(&lpadata.Lpa{
						Donor: lpadata.Donor{
							LastName: "Smith",
						},
					}, nil)

				return lpaStoreResolvingService
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
			err := EnterAccessCodeOptOut(nil, tc.shareCodeStore(), tc.sessionStore(), tc.lpaStoreResolvingService(), actor.TypeAttorney, PathDashboard)(testAppData, w, r)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostEnterAccessCodeOptOutOnShareCodeStoreNotFoundError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd 1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id"))}, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCodeOptOut(template.Execute, shareCodeStore, nil, nil, actor.TypeAttorney, PathDashboard)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
