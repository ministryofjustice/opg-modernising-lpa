package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

	err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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

	err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostEnterVoucher(t *testing.T) {
	testCases := map[string]struct {
		form    url.Values
		voucher actor.Voucher
	}{
		"valid": {
			form: url.Values{
				"first-names": {"John"},
				"last-name":   {"Doe"},
				"email":       {"john@example.com"},
			},
			voucher: actor.Voucher{
				FirstNames: "John",
				LastName:   "Doe",
				Email:      "john@example.com",
			},
		},
		"name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"jane@example.com"},
				"ignore-name-warning": {actor.NewSameNameWarning(actor.TypeVoucher, actor.TypeDonor, "Jane", "Doe").String()},
			},
			voucher: actor.Voucher{
				FirstNames: "Jane",
				LastName:   "Doe",
				Email:      "jane@example.com",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID:   "lpa-id",
					Donor:   actor.Donor{FirstNames: "Jane", LastName: "Doe"},
					Voucher: tc.voucher,
				}).
				Return(nil)

			err := EnterVoucher(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterVoucherWhenInputRequired(t *testing.T) {
	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *enterVoucherData) bool
	}{
		"validation error": {
			form: url.Values{
				"last-name": {"Doe"},
				"email":     {"john@example.com"},
			},
			dataMatcher: func(t *testing.T, data *enterVoucherData) bool {
				return assert.Nil(t, data.NameWarning) &&
					assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
			},
		},
		"name warning": {
			form: url.Values{
				"first-names": {"Jane"},
				"last-name":   {"Doe"},
				"email":       {"john@example.com"},
			},
			dataMatcher: func(t *testing.T, data *enterVoucherData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeVoucher, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
		"other name warning ignored": {
			form: url.Values{
				"first-names":         {"Jane"},
				"last-name":           {"Doe"},
				"email":               {"john@example.com"},
				"ignore-name-warning": {"errorDonorMatchesActor|theVoucher|John|Doe"},
			},
			dataMatcher: func(t *testing.T, data *enterVoucherData) bool {
				return assert.Equal(t, actor.NewSameNameWarning(actor.TypeVoucher, actor.TypeDonor, "Jane", "Doe"), data.NameWarning) &&
					assert.True(t, data.Errors.None())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *enterVoucherData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := EnterVoucher(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				Donor: actor.Donor{FirstNames: "Jane", LastName: "Doe"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
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

	err := EnterVoucher(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

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

func TestVoucherMatches(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
		},
	}

	assert.Equal(t, actor.TypeNone, voucherMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, voucherMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, voucherMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, voucherMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, voucherMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, voucherMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, voucherMatches(donor, "k", "L"))
	assert.Equal(t, actor.TypePersonToNotify, voucherMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, voucherMatches(donor, "o", "p"))
}

func TestVoucherMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &actor.DonorProvidedDetails{
		Donor: actor.Donor{FirstNames: "", LastName: ""},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: actor.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, voucherMatches(donor, "", ""))
}
