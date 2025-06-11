package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAddAnLPA(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, addAnLPAData{
			App:  appcontext.Data{},
			Form: &addAnLPAForm{Options: actor.TypeValues},
		}).
		Return(nil)

	err := AddAnLPA(template.Execute)(appcontext.Data{}, w, r)

	assert.Nil(t, err)
}

func TestGetAddAnLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := AddAnLPA(template.Execute)(appcontext.Data{}, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostAddAnLPA(t *testing.T) {
	testcases := map[actor.Type]Path{
		actor.TypeDonor:                       PathEnterAccessCode,
		actor.TypeCertificateProvider:         PathCertificateProviderEnterReferenceNumber,
		actor.TypeAttorney:                    PathAttorneyEnterReferenceNumber,
		actor.TypeReplacementAttorney:         PathAttorneyEnterReferenceNumber,
		actor.TypeTrustCorporation:            PathAttorneyEnterReferenceNumber,
		actor.TypeReplacementTrustCorporation: PathAttorneyEnterReferenceNumber,
		actor.TypeVoucher:                     PathVoucherEnterReferenceNumber,
		actor.TypePersonToNotify:              PathDashboard,
	}

	for actorType, path := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			f := url.Values{"code-type": {actorType.String()}}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			err := AddAnLPA(nil)(appcontext.Data{}, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, path.String(), resp.Header.Get("Location"))
		})
	}
}

func TestReadAddAnLPAForm(t *testing.T) {
	testcases := []actor.Type{
		actor.TypeDonor,
		actor.TypeCertificateProvider,
		actor.TypeAttorney,
		actor.TypeReplacementAttorney,
		actor.TypeTrustCorporation,
		actor.TypeReplacementTrustCorporation,
		actor.TypeVoucher,
		actor.TypePersonToNotify,
	}

	for _, actorType := range testcases {
		body := url.Values{"code-type": {actorType.String()}}.Encode()

		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r.Header.Add("Content-Type", FormUrlEncoded)

		assert.Equal(t, &addAnLPAForm{
			Options:   actor.TypeValues,
			actorType: actorType,
		}, readAddAnLPAForm(r))
	}
}

func TestValidateAddAnLPAForm(t *testing.T) {
	testcases := map[actor.Type]validation.List{
		actor.TypeDonor: nil,
		actor.TypeNone:  validation.With("code-type", validation.CustomError{Label: "youMustSelectATypeOfAccessCodeToContinue"}),
	}

	for actorType, errors := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			f := addAnLPAForm{actorType: actorType}
			assert.Equal(t, errors, f.Validate())
		})
	}
}
