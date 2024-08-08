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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type changeMobileNumberData struct {
	App        appcontext.Data
	Errors     validation.List
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
			data.Form = readChangeMobileNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if actorType == actor.TypeIndependentWitness {
					provided.IndependentWitness.HasNonUKMobile = data.Form.HasNonUKMobile
					if data.Form.HasNonUKMobile {
						provided.IndependentWitness.Mobile = data.Form.NonUKMobile
					} else {
						provided.IndependentWitness.Mobile = data.Form.Mobile
					}
				} else {
					provided.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile
					if data.Form.HasNonUKMobile {
						provided.CertificateProvider.Mobile = data.Form.NonUKMobile
					} else {
						provided.CertificateProvider.Mobile = data.Form.Mobile
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
	Mobile         string
	HasNonUKMobile bool
	NonUKMobile    string
}

func readChangeMobileNumberForm(r *http.Request) *changeMobileNumberForm {
	return &changeMobileNumberForm{
		Mobile:         page.PostFormString(r, "mobile"),
		HasNonUKMobile: page.PostFormString(r, "has-non-uk-mobile") == "1",
		NonUKMobile:    page.PostFormString(r, "non-uk-mobile"),
	}
}

func (f *changeMobileNumberForm) Validate() validation.List {
	var errors validation.List

	if f.HasNonUKMobile {
		errors.String("non-uk-mobile", "aMobilePhoneNumber", f.NonUKMobile,
			validation.Empty(),
			validation.NonUKMobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	} else {
		errors.String("mobile", "aUKMobileNumber", f.Mobile,
			validation.Empty(),
			validation.Mobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	}

	return errors
}
