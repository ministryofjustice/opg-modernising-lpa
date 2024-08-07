package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterVoucher(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterVoucherData{
			App:  testAppData,
			Form: &enterVoucherForm{},
		}).
		Return(nil)

	err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterVoucherWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostEnterVoucher(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Bloggs"},
		"email":       {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			CertificateProvider: donordata.CertificateProvider{FirstNames: "Barry", LastName: "Bloggs"},
			Voucher: donordata.Voucher{
				FirstNames: "John",
				LastName:   "Bloggs",
				Email:      "john@example.com",
				Allowed:    true,
			},
		}).
		Return(nil)

	err := EnterVoucher(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID:               "lpa-id",
		Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "Barry", LastName: "Bloggs"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCheckYourDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterVoucherWhenMatches(t *testing.T) {
	testcases := map[string]struct{ First, Last string }{
		"donor full name": {"Jane", "Doe"},
		"donor last name": {"John", "Doe"},
		"other full name": {"Barry", "Bloggs"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"first-names": {tc.First},
				"last-name":   {tc.Last},
				"email":       {"jane@example.com"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:               "lpa-id",
					Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
					CertificateProvider: donordata.CertificateProvider{FirstNames: "Barry", LastName: "Bloggs"},
					Voucher: donordata.Voucher{
						FirstNames: tc.First,
						LastName:   tc.Last,
						Email:      "jane@example.com",
					},
				}).
				Return(nil)

			err := EnterVoucher(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:               "lpa-id",
				Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Barry", LastName: "Bloggs"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathConfirmPersonAllowedToVouch.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterVoucherWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"last-name": {"Doe"},
		"email":     {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterVoucherData) bool {
			return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
		})).
		Return(nil)

	err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterVoucherWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"first-names": {"John"},
		"last-name":   {"Doe"},
		"email":       {"john@example.com"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterVoucher(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadEnterVoucherForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"first-names": {"  John "},
		"last-name":   {"Doe"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterVoucherForm(r)

	assert.Equal("John", result.FirstNames)
	assert.Equal("Doe", result.LastName)
}

func TestEnterVoucherFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterVoucherForm
		errors validation.List
	}{
		"valid": {
			form: &enterVoucherForm{
				FirstNames: "A",
				LastName:   "B",
				Email:      "john@example.com",
			},
		},
		"max length": {
			form: &enterVoucherForm{
				FirstNames: strings.Repeat("x", 53),
				LastName:   strings.Repeat("x", 61),
				Email:      "john@example.com",
			},
		},
		"missing all": {
			form: &enterVoucherForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"invalid": {
			form: &enterVoucherForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Email:      "john",
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}).
				With("email", validation.EmailError{Label: "email"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
