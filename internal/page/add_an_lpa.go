package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

func AddAnLPA(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := addAnLPAData{
			App:  appData,
			Form: &addAnLPAForm{Options: actor.TypeValues},
		}

		if r.Method == http.MethodPost {
			data.Form = readAddAnLPAForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.actorType {
				case actor.TypeDonor:
					http.Redirect(w, r, PathEnterAccessCode.Format(), http.StatusFound)
				case actor.TypeCertificateProvider:
					http.Redirect(w, r, PathCertificateProviderEnterReferenceNumber.Format(), http.StatusFound)
				case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
					http.Redirect(w, r, PathAttorneyEnterReferenceNumber.Format(), http.StatusFound)
				case actor.TypeVoucher:
					http.Redirect(w, r, PathVoucherEnterReferenceNumber.Format(), http.StatusFound)
				default:
					http.Redirect(w, r, PathDashboard.Format(), http.StatusFound)
				}
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type addAnLPAData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *addAnLPAForm
}

type addAnLPAForm struct {
	Options   actor.TypeOptions
	actorType actor.Type
}

func readAddAnLPAForm(r *http.Request) *addAnLPAForm {
	actorType, _ := actor.ParseType(PostFormString(r, "code-type"))

	return &addAnLPAForm{
		Options:   actor.TypeValues,
		actorType: actorType,
	}
}

func (f *addAnLPAForm) Validate() validation.List {
	var errors validation.List

	if f.actorType.IsNone() {
		errors.Add("code-type", validation.CustomError{Label: "youMustSelectATypeOfAccessCodeToContinue"})
	}

	return errors
}
