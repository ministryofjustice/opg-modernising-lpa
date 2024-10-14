package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourDeclaration(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{
		Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
	}
	provided := &voucherdata.Provided{LpaID: "lpa-id"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourDeclarationData{
			App:     testAppData,
			Lpa:     lpa,
			Voucher: provided,
			Form:    &yourDeclarationForm{},
		}).
		Return(nil)

	err := YourDeclaration(template.Execute, lpaStoreResolvingService, nil, nil, nil, "")(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourDeclarationWhenSigned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourDeclaration(nil, nil, nil, nil, nil, "")(testAppData, w, r, &voucherdata.Provided{
		LpaID:    "lpa-id",
		SignedAt: time.Now(),
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathThankYou.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetYourDeclarationWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := YourDeclaration(nil, lpaStoreResolvingService, nil, nil, nil, "")(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetYourDeclarationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourDeclaration(template.Execute, lpaStoreResolvingService, nil, nil, nil, "")(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostYourDeclaration(t *testing.T) {
	f := url.Values{
		"confirm": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	testcases := map[string]struct {
		lpa         *lpadata.Lpa
		setupNotify func(*mockNotifyClient)
	}{
		"email": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", ContactLanguagePreference: localize.En},
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorEmail(r.Context(), localize.En, "blah@example.com", "lpa-uid", notify.VoucherHasConfirmedDonorIdentityEmail{
						DonorFullName:     "John Smith",
						DonorStartPageURL: "app:///start",
						VoucherFullName:   "Vivian Voucher",
					}).
					Return(nil)
			},
		},
		"email when signed": {
			lpa: &lpadata.Lpa{
				LpaUID:                           "lpa-uid",
				Donor:                            lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", ContactLanguagePreference: localize.Cy},
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorEmail(r.Context(), localize.Cy, "blah@example.com", "lpa-uid", notify.VoucherHasConfirmedDonorIdentityOnSignedLpaEmail{
						DonorFullName:     "John Smith",
						DonorStartPageURL: "app:///start",
						VoucherFullName:   "Vivian Voucher",
					}).
					Return(nil)
			},
		},
		"mobile": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", Mobile: "0777", ContactLanguagePreference: localize.Cy},
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorSMS(r.Context(), localize.Cy, "0777", "lpa-uid", notify.VoucherHasConfirmedDonorIdentitySMS{
						DonorFullName:     "John Smith",
						DonorStartPageURL: "app:///start",
						VoucherFullName:   "Vivian Voucher",
					}).
					Return(nil)
			},
		},
		"mobile when signed": {
			lpa: &lpadata.Lpa{
				LpaUID:                           "lpa-uid",
				Donor:                            lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", Mobile: "0777", ContactLanguagePreference: localize.En},
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorSMS(r.Context(), localize.En, "0777", "lpa-uid", notify.VoucherHasConfirmedDonorIdentityOnSignedLpaSMS{
						DonorStartPageURL: "app:///start",
						VoucherFullName:   "Vivian Voucher",
					}).
					Return(nil)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			voucherStore := newMockVoucherStore(t)
			voucherStore.EXPECT().
				Put(r.Context(), &voucherdata.Provided{
					LpaID:      "lpa-id",
					FirstNames: "Vivian",
					LastName:   "Voucher",
					SignedAt:   testNow,
					Tasks:      voucherdata.Tasks{SignTheDeclaration: task.StateCompleted},
				}).
				Return(nil)

			notifyClient := newMockNotifyClient(t)
			tc.setupNotify(notifyClient)

			err := YourDeclaration(nil, lpaStoreResolvingService, voucherStore, notifyClient, testNowFn, "app://")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id", FirstNames: "Vivian", LastName: "Voucher"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, voucher.PathThankYou.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourDeclarationWhenValidationError(t *testing.T) {
	f := url.Values{
		"confirm": {"2"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(d *yourDeclarationData) bool {
			return assert.Equal(t, validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToVouch"}), d.Errors)
		})).
		Return(nil)

	err := YourDeclaration(template.Execute, lpaStoreResolvingService, nil, nil, nil, "")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourDeclarationWhenNotifyClientErrors(t *testing.T) {
	f := url.Values{
		"confirm": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	testcases := map[string]struct {
		lpa         *lpadata.Lpa
		setupNotify func(*mockNotifyClient)
	}{
		"email": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com"},
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
		},
		"email when signed": {
			lpa: &lpadata.Lpa{
				LpaUID:   "lpa-uid",
				Donor:    lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com"},
				SignedAt: time.Now(),
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
		},
		"mobile": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Donor:  lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", Mobile: "0777"},
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
		},
		"mobile when signed": {
			lpa: &lpadata.Lpa{
				LpaUID:   "lpa-uid",
				Donor:    lpadata.Donor{FirstNames: "John", LastName: "Smith", Email: "blah@example.com", Mobile: "0777"},
				SignedAt: time.Now(),
			},
			setupNotify: func(m *mockNotifyClient) {
				m.EXPECT().
					SendActorSMS(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			notifyClient := newMockNotifyClient(t)
			tc.setupNotify(notifyClient)

			err := YourDeclaration(nil, lpaStoreResolvingService, nil, notifyClient, testNowFn, "app://")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id", FirstNames: "Vivian", LastName: "Voucher"})

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestPostYourDeclarationWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"confirm": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourDeclaration(nil, lpaStoreResolvingService, voucherStore, notifyClient, testNowFn, "")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestReadYourDeclarationForm(t *testing.T) {
	form := url.Values{
		"confirm": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourDeclarationForm(r)
	assert.Equal(t, true, result.Confirm)
}

func TestYourDeclarationFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourDeclarationForm
		errors validation.List
	}{
		"valid": {
			form: &yourDeclarationForm{
				Confirm: true,
			},
		},
		"not selected": {
			form:   &yourDeclarationForm{},
			errors: validation.With("confirm", validation.CustomError{Label: "youMustSelectTheBoxToVouch"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
