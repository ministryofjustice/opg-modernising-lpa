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
	shareCodeData := actor.ShareCodeData{PK: dynamo.ShareKey(dynamo.DonorShareKey("1"))}

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

			err := DonorAccess(template.Execute, donorStore, shareCodeStore, nil, "", nil)(testLpaAppData, w, r, nil, nil)
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

	err := DonorAccess(nil, donorStore, nil, nil, "", nil)(testLpaAppData, w, r, nil, nil)
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

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "", nil)(testLpaAppData, w, r, nil, nil)
	assert.Equal(t, expectedError, err)
}

func TestPostDonorAccess(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorUID := actoruid.New()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{
			Type:  actor.LpaTypePropertyAndAffairs,
			Donor: actor.Donor{UID: donorUID, FirstNames: "Barry", LastName: "Boy"},
		}, nil)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			Type:  actor.LpaTypePropertyAndAffairs,
			Donor: actor.Donor{UID: donorUID, FirstNames: "Barry", LastName: "Boy", Email: "email@example.com"},
		}).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})
	shareCodeStore.EXPECT().
		PutDonor(r.Context(), testRandomString, actor.ShareCodeData{
			LpaOwnerKey:  dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
			LpaKey:       dynamo.LpaKey("lpa-id"),
			ActorUID:     donorUID,
			InviteSentTo: "email@example.com",
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), "email@example.com", notify.DonorAccessEmail{
			SupporterFullName: "John Smith",
			OrganisationName:  "Helpers",
			LpaType:           "translation",
			DonorName:         "Barry Boy",
			URL:               "http://whatever/start",
			ShareCode:         testRandomString,
		}).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(actor.LpaTypePropertyAndAffairs.String()).
		Return("Translation")
	testLpaAppData.Localizer = localizer

	err := DonorAccess(nil, donorStore, shareCodeStore, notifyClient, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{PK: dynamo.OrganisationKey("org-id"), ID: "org-id", Name: "Helpers"}, &actor.Member{FirstNames: "John", LastName: "Smith"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ViewLPA.Format("lpa-id")+"?inviteSentTo=email%40example.com", resp.Header.Get("Location"))
}

func TestPostDonorAccessWhenDonorUpdateErrors(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "", nil)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"}, nil)
	assert.Equal(t, expectedError, err)
}

func TestPostDonorAccessWhenShareCodeStoreErrors(t *testing.T) {
	form := url.Values{"email": {"email@example.com"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{Donor: actor.Donor{Email: "email@example.com"}}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(actor.ShareCodeData{}, dynamo.NotFoundError{})
	shareCodeStore.EXPECT().
		PutDonor(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"}, nil)
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
		Return(&actor.DonorProvidedDetails{Donor: actor.Donor{Email: "email@example.com"}}, nil)

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

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("Translation")
	testLpaAppData.Localizer = localizer

	err := DonorAccess(nil, donorStore, shareCodeStore, notifyClient, "", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"}, &actor.Member{})
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

	err := DonorAccess(template.Execute, donorStore, shareCodeStore, nil, "", nil)(testLpaAppData, w, r, &actor.Organisation{ID: "org-id"}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDonorAccessRecall(t *testing.T) {
	form := url.Values{"action": {"recall"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeData := actor.ShareCodeData{PK: dynamo.ShareKey(dynamo.DonorShareKey("1")), InviteSentTo: "email@example.com"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(shareCodeData, nil)
	shareCodeStore.EXPECT().
		Delete(r.Context(), shareCodeData).
		Return(nil)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ViewLPA.Format("lpa-id")+"?inviteRecalledFor=email%40example.com", resp.Header.Get("Location"))
}

func TestPostDonorAccessRecallWhenDeleteErrors(t *testing.T) {
	form := url.Values{"action": {"recall"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeData := actor.ShareCodeData{PK: dynamo.ShareKey(dynamo.DonorShareKey("1")), InviteSentTo: "email@example.com"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(shareCodeData, nil)
	shareCodeStore.EXPECT().
		Delete(r.Context(), shareCodeData).
		Return(expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{}, &actor.Member{})
	assert.Equal(t, expectedError, err)
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

func TestPostDonorAccessRemove(t *testing.T) {
	form := url.Values{"action": {"remove"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeData := actor.ShareCodeData{
		PK:           dynamo.ShareKey(dynamo.DonorShareKey("1")),
		SK:           dynamo.ShareSortKey(dynamo.DonorInviteKey("donor-session-id", "lpa-id")),
		InviteSentTo: "email@example.com",
		LpaOwnerKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id")),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(shareCodeData, nil)

	donor := &actor.DonorProvidedDetails{SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id"))}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(donor, nil)
	donorStore.EXPECT().
		DeleteDonorAccess(r.Context(), shareCodeData).
		Return(nil)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ViewLPA.Format("lpa-id")+"?accessRemovedFor=email%40example.com", resp.Header.Get("Location"))
}

func TestPostDonorAccessRemoveWhenDonorHasPaid(t *testing.T) {
	form := url.Values{"action": {"remove"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeData := actor.ShareCodeData{
		PK:           dynamo.ShareKey(dynamo.DonorShareKey("1")),
		SK:           dynamo.ShareSortKey(dynamo.DonorInviteKey("donor-session-id", "lpa-id")),
		InviteSentTo: "email@example.com",
		LpaOwnerKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id")),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(r.Context()).
		Return(shareCodeData, nil)

	donor := &actor.DonorProvidedDetails{SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDonorAccessRemoveWhenDeleteLinkError(t *testing.T) {
	form := url.Values{"action": {"remove"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		GetDonor(mock.Anything).
		Return(actor.ShareCodeData{}, nil)

	donor := &actor.DonorProvidedDetails{SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id"))}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)
	donorStore.EXPECT().
		DeleteDonorAccess(mock.Anything, mock.Anything).
		Return(expectedError)

	err := DonorAccess(nil, donorStore, shareCodeStore, nil, "http://whatever", testRandomStringFn)(testLpaAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
