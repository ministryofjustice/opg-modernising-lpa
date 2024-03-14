package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDonorAccess(t *testing.T) {
	donor := &actor.DonorProvidedDetails{Donor: actor.Donor{Email: "x"}}
	shareCodeData := actor.ShareCodeData{PK: "1"}

	testcases := map[string]struct {
		data                *donorAccessData
		shareCodeStoreError error
	}{
		"not sent": {
			data: &donorAccessData{
				App:   testLpaAppData,
				Donor: donor,
				Form:  &donorAccessForm{Email: "x"},
			},
			shareCodeStoreError: dynamo.NotFoundError{},
		},
		"sent": {
			data: &donorAccessData{
				App:       testLpaAppData,
				Donor:     donor,
				Form:      &donorAccessForm{Email: "x"},
				ShareCode: &shareCodeData,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(r.Context()).
				Return(donor, nil)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				GetDonor(r.Context()).
				Return(shareCodeData, tc.shareCodeStoreError)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(expectedError)

			err := DonorAccess(template.Execute, donorStore, shareCodeStore, nil, nil)(testLpaAppData, w, r, nil)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetDonorAccessWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := DonorAccess(nil, donorStore, nil, nil, nil)(testLpaAppData, w, r, nil)
	assert.Equal(t, expectedError, err)
}

func TestGetDonorAccessWhenShareCodeStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, nil)(testLpaAppData, w, r, nil)
	assert.Equal(t, expectedError, err)
}

func TestPostDonorAccess(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorUID := actoruid.New()
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{UID: donorUID},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})
	shareCodeStore.EXPECT().
		PutDonor(r.Context(), testRandomString, actor.ShareCodeData{
			SessionID:    "org-id",
			LpaID:        "lpa-id",
			ActorUID:     donorUID,
			InviteSentTo: "email@example.com",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), "email@example.com", notify.DonorAccessEmail{ShareCode: testRandomString}).
		Return(nil)

	err := DonorAccess(nil, donorStore, shareCodeStore, notifyClient, testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ViewLPA.Format()+"?id=lpa-id&inviteSentTo=email%40example.com", resp.Header.Get("Location"))
}

func TestPostDonorAccessWhenShareCodeStoreErrors(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})
	shareCodeStore.EXPECT().
		PutDonor(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostDonorAccessWhenNotifyErrors(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})
	shareCodeStore.EXPECT().
		PutDonor(r.Context(), mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, notifyClient, testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostDonorAccessWhenValidationError(t *testing.T) {
	f := url.Values{"email": {"x"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *donorAccessData) bool {
			return assert.Equal(t, validation.With("email", validation.EmailError{Label: "email"}), data.Errors)
		})).
		Return(nil)

	err := DonorAccess(template.Execute, donorStore, shareCodeStore, nil, nil)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadDonorAccessForm(t *testing.T) {
	form := url.Values{
		"email": {"email@example.com"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readDonorAccessForm(r)

	assert.Equal(t, "email@example.com", result.Email)
}

func TestDonorAccessFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *donorAccessForm
		errors validation.List
	}{
		"valid": {
			form: &donorAccessForm{
				Email: "email@example.com",
			},
		},
		"missing": {
			form:   &donorAccessForm{},
			errors: validation.With("email", validation.EnterError{Label: "email"}),
		},
		"invalid": {
			form: &donorAccessForm{
				Email: "x",
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
