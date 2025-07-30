package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(r.Context(), "lpa-uid").
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:     testAppData,
			Form:    &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			Options: lpadata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaUID: "lpa-uid",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenCertificateProviderHasLinked(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(r.Context(), "lpa-uid").
		Return(&certificateproviderdata.Provided{}, nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 testAppData,
			CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper},
			Form:                &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{CarryOutBy: lpadata.ChannelPaper},
			Options:             lpadata.ChannelValues,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:     testAppData,
			Form:    &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			Options: lpadata.ChannelValues,
		}).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRole(t *testing.T) {
	testCases := []struct {
		carryOutBy lpadata.Channel
		email      string
	}{
		{
			carryOutBy: lpadata.ChannelPaper,
		},
		{
			carryOutBy: lpadata.ChannelOnline,
			email:      "someone@example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.carryOutBy.String(), func(t *testing.T) {
			form := url.Values{
				"carry-out-by": {tc.carryOutBy.String()},
				"email":        {tc.email},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			certificateProvider := donordata.CertificateProvider{CarryOutBy: tc.carryOutBy, Email: tc.email}

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				OneByUID(mock.Anything, mock.Anything).
				Return(nil, dynamo.NotFoundError{})

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutCertificateProvider(r.Context(), certificateProvider).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					CertificateProvider: donordata.CertificateProvider{CarryOutBy: tc.carryOutBy, Email: tc.email},
				}).
				Return(nil)

			err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore, certificateProviderStore, reuseStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathCertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleChangingFromOnlineToPaper(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelPaper.String()},
		"email":        {"a@b.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProvider := donordata.CertificateProvider{CarryOutBy: lpadata.ChannelPaper}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(r.Context(), certificateProvider).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			CertificateProvider: certificateProvider,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore, certificateProviderStore, reuseStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:               "lpa-id",
		CertificateProvider: donordata.CertificateProvider{CarryOutBy: lpadata.ChannelOnline, Email: "a@b.com"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenEmailChanged(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelOnline.String()},
		"email":        {"b@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invitedAt := time.Now()
	uid := actoruid.New()
	updatedCertificateProvider := donordata.CertificateProvider{
		UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "b@example.com",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(r.Context(), updatedCertificateProvider).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			CertificateProvider: donordata.CertificateProvider{
				UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "b@example.com",
			},
			CertificateProviderInvitedAt: testNow,
		}).
		Return(nil)

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(r.Context(), uid).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendCertificateProviderInvite(r.Context(), testAppData, &donordata.Provided{
			LpaID: "lpa-id",
			CertificateProvider: donordata.CertificateProvider{
				UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "b@example.com",
			},
			CertificateProviderInvitedAt: invitedAt,
		}).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore, certificateProviderStore, reuseStore, accessCodeStore, accessCodeSender, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		CertificateProvider: donordata.CertificateProvider{
			UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "a@example.com",
		},
		CertificateProviderInvitedAt: invitedAt,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCertificateProviderAddress.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenEmailChangedWhenDeleteByActorError(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelOnline.String()},
		"email":        {"b@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invitedAt := time.Now()
	uid := actoruid.New()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, nil, certificateProviderStore, nil, accessCodeStore, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		CertificateProvider: donordata.CertificateProvider{
			UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "a@example.com",
		},
		CertificateProviderInvitedAt: invitedAt,
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenEmailChangedWhenSendCertificateProviderInviteError(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelOnline.String()},
		"email":        {"b@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invitedAt := time.Now()
	uid := actoruid.New()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	accessCodeStore := newMockAccessCodeStore(t)
	accessCodeStore.EXPECT().
		DeleteByActor(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendCertificateProviderInvite(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, nil, certificateProviderStore, nil, accessCodeStore, accessCodeSender, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		CertificateProvider: donordata.CertificateProvider{
			UID: uid, CarryOutBy: lpadata.ChannelOnline, Email: "a@example.com",
		},
		CertificateProviderInvitedAt: invitedAt,
	})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenReuseStoreErrors(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelPaper.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(mock.Anything, mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, nil, certificateProviderStore, reuseStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})

	assert.ErrorIs(t, err, expectedError)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenDonorStoreErrors(t *testing.T) {
	form := url.Values{
		"carry-out-by": {lpadata.ChannelPaper.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutCertificateProvider(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(nil, donorStore, certificateProviderStore, reuseStore, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostHowWouldCertificateProviderPreferToCarryOutTheirRoleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howWouldCertificateProviderPreferToCarryOutTheirRoleData) bool {
			return assert.Equal(t, validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}), data.Errors)
		})).
		Return(nil)

	err := HowWouldCertificateProviderPreferToCarryOutTheirRole(template.Execute, nil, certificateProviderStore, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadHowWouldCertificateProviderPreferToCarryOutTheirRoleForm(t *testing.T) {
	testcases := map[string]struct {
		carryOutBy   lpadata.Channel
		email        string
		formValues   url.Values
		expectedForm *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	}{
		"online with email": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelOnline.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "a@b.com",
			},
		},
		"paper": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelPaper.String()},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
		"paper with email": {
			formValues: url.Values{
				"carry-out-by": {lpadata.ChannelPaper.String()},
				"email":        {"a@b.com"},
			},
			expectedForm: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.formValues.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			result := readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)

			assert.Equal(t, tc.expectedForm, result)
		})
	}
}

func TestHowWouldCertificateProviderPreferToCarryOutTheirRoleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
		errors validation.List
	}{
		"paper": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelPaper,
			},
		},
		"online": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "someone@example.com",
			},
		},
		"online email invalid": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
				Email:      "what",
			},
			errors: validation.With("email", validation.EmailError{Label: "certificateProvidersEmail"}),
		},
		"online email missing": {
			form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpadata.ChannelOnline,
			},
			errors: validation.With("email", validation.EnterError{Label: "certificateProvidersEmail"}),
		},
		"missing": {
			form:   &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{},
			errors: validation.With("carry-out-by", validation.SelectError{Label: "howYourCertificateProviderWouldPreferToCarryOutTheirRole"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
