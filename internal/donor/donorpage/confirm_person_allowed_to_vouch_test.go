package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmPersonAllowedToVouchDataMultipleMatches(t *testing.T) {
	testcases := map[*confirmPersonAllowedToVouchData]bool{
		{}:                                       false,
		{Matches: []actor.Type{actor.TypeDonor}}: false,
		{MatchSurname: true}:                     false,
		{Matches: []actor.Type{actor.TypeDonor, actor.TypeDonor}}:    true,
		{Matches: []actor.Type{actor.TypeDonor}, MatchSurname: true}: true,
	}

	for data, expected := range testcases {
		assert.Equal(t, expected, data.MultipleMatches())
	}
}

func TestGetConfirmPersonAllowedToVouch(t *testing.T) {
	testcases := map[string]struct {
		donor        *donordata.Provided
		matches      []actor.Type
		matchSurname bool
	}{
		"matches donor": {
			donor: &donordata.Provided{
				Donor:               donordata.Donor{FirstNames: "John", LastName: "Smith"},
				CertificateProvider: donordata.CertificateProvider{FirstNames: "John", LastName: "Smith"},
				Voucher:             donordata.Voucher{FirstNames: "John", LastName: "Smith"},
			},
			matches: []actor.Type{actor.TypeDonor, actor.TypeCertificateProvider},
		},
		"matches donor last name": {
			donor: &donordata.Provided{
				Donor:               donordata.Donor{FirstNames: "Dave", LastName: "Smith"},
				CertificateProvider: donordata.CertificateProvider{FirstNames: "John", LastName: "Smith"},
				Voucher:             donordata.Voucher{FirstNames: "John", LastName: "Smith"},
			},
			matches:      []actor.Type{actor.TypeCertificateProvider},
			matchSurname: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &confirmPersonAllowedToVouchData{
					App:          testAppData,
					Form:         form.NewYesNoForm(form.YesNoUnknown),
					Matches:      tc.matches,
					MatchSurname: tc.matchSurname,
					FullName:     "John Smith",
				}).
				Return(nil)

			err := ConfirmPersonAllowedToVouch(template.Execute, nil)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmPersonAllowedToVouchWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmPersonAllowedToVouch(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmPersonAllowedToVouch(t *testing.T) {
	testCases := map[string]struct {
		yesNo    form.YesNo
		voucher  donordata.Voucher
		redirect donor.Path
	}{
		"yes": {
			yesNo:    form.Yes,
			voucher:  donordata.Voucher{FirstNames: "John", Allowed: true},
			redirect: donor.PathCheckYourDetails,
		},
		"no": {
			yesNo:    form.No,
			redirect: donor.PathEnterVoucher,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:   "lpa-id",
					Voucher: tc.voucher,
				}).
				Return(nil)

			err := ConfirmPersonAllowedToVouch(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID:   "lpa-id",
				Voucher: donordata.Voucher{FirstNames: "John"},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostConfirmPersonAllowedToVouchWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmPersonAllowedToVouch(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmPersonAllowedToVouchWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *confirmPersonAllowedToVouchData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesIfPersonIsAllowedToVouchForYou"}), data.Errors)
		})).
		Return(nil)

	err := ConfirmPersonAllowedToVouch(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
