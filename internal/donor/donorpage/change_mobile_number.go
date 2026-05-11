package donorpage

import (
	"context"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type changeMobileNumberData struct {
	App        appcontext.Data
	Form       *changeMobileNumberForm
	ActorType  actor.Type
	FirstNames string
}

func ChangeMobileNumber(tmpl template.Template, witnessCodeSender WitnessCodeSender, actorType actor.Type) Handler {
	var send func(context.Context, *donordata.Provided) error
	var redirect donor.Path
	switch actorType {
	case actor.TypeIndependentWitness:
		send = witnessCodeSender.SendToIndependentWitness
		redirect = donor.PathWitnessingAsIndependentWitness
	case actor.TypeCertificateProvider:
		send = witnessCodeSender.SendToCertificateProvider
		redirect = donor.PathWitnessingAsCertificateProvider
	default:
		panic("ChangeMobileNumber only supports IndependentWitness or CertificateProvider actors")
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &changeMobileNumberData{
			App:        appData,
			Form:       &changeMobileNumberForm{},
			ActorType:  actorType,
			FirstNames: provided.CertificateProvider.FirstNames,
		}

		if actorType == actor.TypeIndependentWitness {
			data.FirstNames = provided.IndependentWitness.FirstNames
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				if actorType == actor.TypeIndependentWitness {
					provided.IndependentWitness.HasNonUKMobile = data.Form.HasNonUKMobile.Value
					if data.Form.HasNonUKMobile.Value {
						provided.IndependentWitness.Mobile = data.Form.NonUKMobile.Value
					} else {
						provided.IndependentWitness.Mobile = data.Form.Mobile.Value
					}
				} else {
					provided.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile.Value
					if data.Form.HasNonUKMobile.Value {
						provided.CertificateProvider.Mobile = data.Form.NonUKMobile.Value
					} else {
						provided.CertificateProvider.Mobile = data.Form.Mobile.Value
					}
				}

				if err := send(r.Context(), provided); err != nil {
					if errors.Is(err, donor.ErrTooManyWitnessCodeRequests) {
						data.Errors.Add("request", validation.CustomError{Label: "pleaseWaitOneMinute"})
						return tmpl(w, data)
					}

					return err
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type changeMobileNumberForm struct {
	newforms.Form
	HasNonUKMobile *newforms.Bool
	Mobile         *newforms.String
	NonUKMobile    *newforms.String
}

func newChangeMobileNumberForm(l Localizer) *changeMobileNumberForm {
	return &changeMobileNumberForm{
		HasNonUKMobile: newforms.NewBool("has-non-uk-mobile", l.T("iDoNotHaveAUkMobileNumber")),
		Mobile:         newforms.NewString("mobile", l.T("aUKMobileNumber")),
		NonUKMobile:    newforms.NewString("non-uk-mobile", l.T("aMobilePhoneNumber")),
	}
}

func (f *changeMobileNumberForm) Parse(r *http.Request) bool {
	ok := f.ParsePostForm(r, f.HasNonUKMobile)

	if f.HasNonUKMobile.Value {
		ok = f.ParsePostForm(r, f.NonUKMobile) && ok
	} else {
		ok = f.ParsePostForm(r, f.Mobile) && ok
	}

	return ok
}
