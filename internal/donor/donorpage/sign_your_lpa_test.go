package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &signYourLpaData{
			App:                  testAppData,
			Form:                 &signYourLpaForm{},
			Donor:                &donordata.Provided{},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Execute, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSignYourLpaWhenSigned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SignYourLpa(nil, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:            "lpa-id",
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		SignedAt:         time.Now(),
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWitnessingYourSignature.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetSignYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		WantToSignLpa:     true,
		WantToApplyForLpa: false,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &signYourLpaData{
			App:   testAppData,
			Donor: donor,
			Form: &signYourLpaForm{
				WantToSign:  true,
				WantToApply: false,
			},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Execute, nil, nil, testNowFn)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSignYourLpa(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:             "lpa-id",
			IdentityUserData:  identity.UserData{Status: identity.StatusConfirmed},
			WantToSignLpa:     true,
			WantToApplyForLpa: true,
			SignedAt:          testNow,
		}).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(r.Context(), scheduled.Event{
			At:     testNow.AddDate(0, 3, 1),
			Action: scheduleddata.ActionRemindCertificateProviderToComplete,
		}, scheduled.Event{
			At:     testNow.AddDate(0, 21, 1),
			Action: scheduleddata.ActionRemindCertificateProviderToComplete,
		}).
		Return(nil)

	err := SignYourLpa(nil, donorStore, scheduledStore, testNowFn)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", IdentityUserData: identity.UserData{Status: identity.StatusConfirmed}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWitnessingYourSignature.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostSignYourLpaWhenScheduledStoreErrors(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := SignYourLpa(nil, nil, scheduledStore, testNowFn)(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostSignYourLpaWhenDonorStoreErrors(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := SignYourLpa(nil, donorStore, scheduledStore, testNowFn)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostSignYourLpaWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"unrecognised-signature", "another-unrecognised-signature"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *signYourLpaData) bool {
			return assert.Equal(t, validation.With("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"}), data.Errors)
		})).
		Return(nil)

	err := SignYourLpa(template.Execute, nil, nil, testNowFn)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadSignYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readSignYourLpaForm(r, localize.Cy)

	assert.Equal(localize.Cy, result.lpaLanguage)
	assert.Equal(true, result.WantToSign)
	assert.Equal(true, result.WantToApply)
}

func TestSignYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *signYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &signYourLpaForm{
				WantToApply: true,
				WantToSign:  true,
			},
		},
		"valid but wrong language": {
			form: &signYourLpaForm{
				WantToApply:   true,
				WantToSign:    true,
				WrongLanguage: true,
				lpaLanguage:   localize.Cy,
			},
			errors: validation.With("sign-lpa", youMustViewAndSignInLanguageError{LpaLanguage: localize.Cy}),
		},
		"only want-to-sign selected": {
			form: &signYourLpaForm{
				WantToSign:    true,
				WrongLanguage: true,
			},
			errors: validation.With("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"}),
		},
		"only want-to-apply selected": {
			form: &signYourLpaForm{
				WantToApply: true,
			},
			errors: validation.With("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"}),
		},
		"none selected": {
			form:   &signYourLpaForm{},
			errors: validation.With("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestYouMustViewAndSignInLanguageError(t *testing.T) {
	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("in:cy").
		Return("in Welsh")
	localizer.EXPECT().
		Format("youMustViewAndSignInLanguage", map[string]any{
			"InLang": "in Welsh",
		}).
		Return("some words")

	assert.Equal(t, "some words", youMustViewAndSignInLanguageError{LpaLanguage: localize.Cy}.Format(localizer))
}
