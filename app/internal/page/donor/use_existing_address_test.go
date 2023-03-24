package donor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetUseExistingAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?subjectId=2&type=attorney", nil)

	subject := actor.Attorney{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress}

	lpa := &page.Lpa{Attorneys: []actor.Attorney{
		{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
		subject,
	}}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, useExistingAddressData{
			App: testAppData,
			Addresses: []page.AddressDetail{
				{Address: testAddress, Role: actor.TypeAttorney, ID: "1", Name: "Joe Smith"},
				{Address: testAddress, Role: actor.TypeAttorney, ID: "2", Name: "Joan Smith"},
			},
			Subject: subject,
		}).
		Return(nil)

	err := UseExistingAddress(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUseExistingAddressStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?subjectId=2&type=attorney", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUseExistingAddressSubjectNotFound(t *testing.T) {
	testCases := map[string]struct {
		Type          string
		ExpectedError error
	}{
		"attorney":             {Type: "attorney", ExpectedError: errors.New("attorney not found")},
		"replacement attorney": {Type: "replacementAttorney", ExpectedError: errors.New("replacementAttorney not found")},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?subjectId=2&type="+tc.Type, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)

			err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.ExpectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetUseExistingAddressNoAddresses(t *testing.T) {
	testCases := map[string]struct {
		Addresses []page.AddressDetail
		Subject   actor.Attorney
	}{
		"no addresses": {
			Addresses: []page.AddressDetail{},
			Subject:   actor.Attorney{ID: "1"},
		},
		"1 address belonging to subject": {
			Addresses: []page.AddressDetail{{ID: "1", Address: testAddress}},
			Subject:   actor.Attorney{ID: "1", Address: testAddress},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?subjectId=1&type=attorney&from=/somewhere", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{Attorneys: actor.Attorneys{tc.Subject}, ID: "lpa-id"}, nil)

			err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id/somewhere", resp.Header.Get("Location"))
		})
	}
}

func TestGetUseExistingAddressTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?subjectId=2&type=attorney", nil)

	subject := actor.Attorney{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress}

	lpa := &page.Lpa{Attorneys: []actor.Attorney{
		{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
		subject,
	}}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, useExistingAddressData{
			App: testAppData,
			Addresses: []page.AddressDetail{
				{Address: testAddress, Role: actor.TypeAttorney, ID: "1", Name: "Joe Smith"},
				{Address: testAddress, Role: actor.TypeAttorney, ID: "2", Name: "Joan Smith"},
			},
			Subject: subject,
		}).
		Return(expectedError)

	err := UseExistingAddress(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostUseExistingAddress(t *testing.T) {
	newAddress := place.Address{Line1: "1 New Road"}

	testCases := map[string]struct {
		Lpa         *page.Lpa
		UpdatedLpa  *page.Lpa
		Type        string
		ExpectedUrl string
	}{
		"subject is attorney": {
			Lpa: &page.Lpa{Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			}},
			UpdatedLpa: &page.Lpa{Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: newAddress},
			}},
			Type:        "attorney",
			ExpectedUrl: "/lpa/lpa-id" + testAppData.Paths.ChooseAttorneysSummary,
		},
		"subject is replacement attorney": {
			Lpa: &page.Lpa{ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			}},
			UpdatedLpa: &page.Lpa{ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: newAddress},
			}},
			Type:        "replacementAttorney",
			ExpectedUrl: "/lpa/lpa-id" + testAppData.Paths.ChooseReplacementAttorneysSummary,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			form := url.Values{
				"address-index": {"0"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/?subjectId=2&type="+tc.Type, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.Lpa, nil)

			lpaStore.
				On("Put", r.Context(), tc.UpdatedLpa).
				Return(nil)

			err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedUrl, resp.Header.Get("Location"))
		})
	}
}

func TestPostUseExistingAddressWithMultipleAddresses(t *testing.T) {
	newAddress := place.Address{Line1: "1 New Road"}

	w := httptest.NewRecorder()
	form := url.Values{
		"address-index": {"0"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?subjectId=2&type=replacementAttorney", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			},
			ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Janet", LastName: "Smith", Address: testAddress},
				{ID: "2", FirstNames: "JoJo", LastName: "Smith", Address: testAddress},
			},
			CertificateProvider: actor.CertificateProvider{FirstNames: "Jorge", LastName: "Smith", Address: newAddress},
		}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			},
			ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Janet", LastName: "Smith", Address: testAddress},
				{ID: "2", FirstNames: "JoJo", LastName: "Smith", Address: newAddress},
			},
			CertificateProvider: actor.CertificateProvider{FirstNames: "Jorge", LastName: "Smith", Address: newAddress},
		}).
		Return(nil)

	err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+testAppData.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostUseExistingAddressStoreError(t *testing.T) {
	newAddress := place.Address{Line1: "1 New Road"}

	testCases := map[string]struct {
		Lpa         *page.Lpa
		UpdatedLpa  *page.Lpa
		Type        string
		ExpectedUrl string
	}{
		"subject is attorney": {
			Lpa: &page.Lpa{Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			}},
			UpdatedLpa: &page.Lpa{Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: newAddress},
			}},
			Type:        "attorney",
			ExpectedUrl: "/lpa/lpa-id" + testAppData.Paths.ChooseAttorneysSummary,
		},
		"subject is replacement attorney": {
			Lpa: &page.Lpa{ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			}},
			UpdatedLpa: &page.Lpa{ReplacementAttorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: newAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: newAddress},
			}},
			Type:        "replacementAttorney",
			ExpectedUrl: "/lpa/lpa-id" + testAppData.Paths.ChooseReplacementAttorneysSummary,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			form := url.Values{
				"address-index": {"0"},
			}

			r, _ := http.NewRequest(http.MethodPost, "/?subjectId=2&type="+tc.Type, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.Lpa, nil)

			lpaStore.
				On("Put", r.Context(), tc.UpdatedLpa).
				Return(expectedError)

			err := UseExistingAddress(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostUseExistingAddressValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"address-index": {"not-expected"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?subjectId=2&type=attorney", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Attorneys: []actor.Attorney{
				{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
				{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, useExistingAddressData{
			App:    testAppData,
			Errors: validation.With("address-index", validation.SelectError{Label: "address"}),
			Addresses: []page.AddressDetail{
				{Address: testAddress, Role: actor.TypeAttorney, ID: "1", Name: "Joe Smith"},
				{Address: testAddress, Role: actor.TypeAttorney, ID: "2", Name: "Joan Smith"},
			},
			Subject: actor.Attorney{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
			Form:    &UseExistingAddressForm{AddressIndex: "not-expected"},
		}).
		Return(nil)

	err := UseExistingAddress(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadUseExistingAddressForm(t *testing.T) {
	vals := url.Values{
		"address-index": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?subjectId=2&type=attorney", strings.NewReader(vals.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	form := readUseExistingAddressForm(r)

	assert.Equal(t, "1", form.AddressIndex)
}

func TestGetSubject(t *testing.T) {
	lpa := &page.Lpa{
		Attorneys: []actor.Attorney{
			{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
			{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
		},
		ReplacementAttorneys: []actor.Attorney{
			{ID: "1", FirstNames: "Jorge", LastName: "Smith", Address: testAddress},
			{ID: "2", FirstNames: "Janet", LastName: "Smith", Address: testAddress},
		},
	}

	attorney, found := getSubject("attorney", "2", lpa)
	assert.True(t, found)
	assert.Equal(t, "Joan", attorney.FirstNames)
	assert.Equal(t, "Smith", attorney.LastName)
	assert.Equal(t, "2", attorney.ID)
	assert.Equal(t, testAddress, attorney.Address)

	replacementAttorney, found := getSubject("replacementAttorney", "2", lpa)
	assert.True(t, found)
	assert.Equal(t, "Janet", replacementAttorney.FirstNames)
	assert.Equal(t, "Smith", replacementAttorney.LastName)
	assert.Equal(t, "2", replacementAttorney.ID)
	assert.Equal(t, testAddress, replacementAttorney.Address)
}

func TestGetSubjectNotFound(t *testing.T) {
	lpa := &page.Lpa{
		Attorneys: []actor.Attorney{
			{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress},
			{ID: "2", FirstNames: "Joan", LastName: "Smith", Address: testAddress},
		},
		ReplacementAttorneys: []actor.Attorney{
			{ID: "1", FirstNames: "Jorge", LastName: "Smith", Address: testAddress},
			{ID: "2", FirstNames: "Janet", LastName: "Smith", Address: testAddress},
		},
	}

	_, found := getSubject("attorney", "3", lpa)
	assert.False(t, found)

	_, found = getSubject("replacementAttorney", "3", lpa)
	assert.False(t, found)
}

func TestAddressDetailsContains(t *testing.T) {
	attorney1 := actor.Attorney{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress}
	attorney2 := actor.Attorney{ID: "1", FirstNames: "Joe", LastName: "Smith", Address: testAddress}
	ad := []page.AddressDetail{
		{Address: testAddress, Role: actor.TypeAttorney, ID: "1", Name: "Joe Smith"},
		{Address: testAddress, Role: actor.TypeAttorney, ID: "2", Name: "Joan Smith"},
	}

	assert.True(t, addressDetailsContains(attorney1, ad))
	assert.True(t, addressDetailsContains(attorney2, ad))
}

func TestAddressDetailsContainsNotFound(t *testing.T) {
	attorney1 := actor.Attorney{ID: "wrong-id", FirstNames: "Joe", LastName: "Smith", Address: testAddress}
	attorney2 := actor.Attorney{ID: "2", FirstNames: "Wrong", LastName: "Name", Address: testAddress}

	ad := []page.AddressDetail{
		{Address: testAddress, Role: actor.TypeAttorney, ID: "1", Name: "Joe Smith"},
		{Address: testAddress, Role: actor.TypeAttorney, ID: "2", Name: "Joan Smith"},
	}

	assert.False(t, addressDetailsContains(attorney1, ad))
	assert.False(t, addressDetailsContains(attorney2, ad))
}
