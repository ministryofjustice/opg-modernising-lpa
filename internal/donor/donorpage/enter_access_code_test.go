package donorpage

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	err := EnterAccessCode(nil, template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

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

	err := EnterAccessCode(nil, template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCode(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	session := &sesh.LoginSession{
		Sub:     "hey",
		Email:   "a@example.com",
		HasLPAs: true,
	}

	accessCode := accesscodedata.DonorLink{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("123")),
		ActorUID:    testUID,
		LpaUID:      "lpa-uid",
	}

	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		Donor:  lpadata.Donor{LastName: "Smith"},
	}

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(r.Context(), accesscodedata.HashedFromString("abcd1234", "Smith")).
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

		return assert.Equal(t, &appcontext.Session{SessionID: "aGV5", LpaID: "lpa-id", OrganisationID: "123"}, session)
	})

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(newCtx).
		Return(lpa, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(r.Context(), accessCode, "a@example.com").
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "donor access added", slog.String("lpa_id", "lpa-id"))

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineDonor).
		Return(nil)

	err := EnterAccessCode(logger, nil, accessCodeStore, sessionStore, lpaStoreResolvingService, donorStore, eventClient)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathDashboard.Format(), resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeWhenDonorLastNameIncorrect(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smithy"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCode := accesscodedata.DonorLink{
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
		GetDonor(mock.Anything, mock.Anything).
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

	err := EnterAccessCode(nil, template.Execute, accessCodeStore, sessionStore, lpaStoreResolvingService, nil, nil)(testAppData, w, r)
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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{}, expectedError)

	err := EnterAccessCode(nil, nil, accessCodeStore, nil, nil, nil, nil)(testAppData, w, r)

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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"}), data.Errors)
		})).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{}, dynamo.NotFoundError{})

	err := EnterAccessCode(nil, template.Execute, accessCodeStore, nil, nil, nil, nil)(testAppData, w, r)

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
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := EnterAccessCode(nil, nil, accessCodeStore, sessionStore, lpaStoreResolvingService, nil, nil)(testAppData, w, r)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnSessionGetError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(r.Context(), accesscodedata.HashedFromString("abcd1234", "Smith")).
		Return(accesscodedata.DonorLink{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, expectedError)

	err := EnterAccessCode(nil, nil, accessCodeStore, sessionStore, nil, nil, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnSessionSetError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

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

	err := EnterAccessCode(nil, nil, accessCodeStore, sessionStore, lpaStoreResolvingService, nil, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeOnValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.AccessCode:    {""},
		form.FieldNames.DonorLastName: {"abc"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(data enterAccessCodeData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.AccessCode, validation.EnterError{Label: "yourAccessCode"}), data.Errors)
		})).
		Return(nil)

	err := EnterAccessCode(nil, template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnDonorStoreError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, nil, accessCodeStore, sessionStore, lpaStoreResolvingService, donorStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnEventClientError(t *testing.T) {
	form := url.Values{
		form.FieldNames.AccessCode:    {"abcd1234"},
		form.FieldNames.DonorLastName: {"Smith"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		GetDonor(mock.Anything, mock.Anything).
		Return(accesscodedata.DonorLink{LpaKey: dynamo.LpaKey("lpa-id")}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(logger, nil, accessCodeStore, sessionStore, lpaStoreResolvingService, donorStore, eventClient)(testAppData, w, r)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
