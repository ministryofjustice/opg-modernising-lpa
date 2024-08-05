package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
	Options             lpadata.CertificateProviderRelationshipOptions
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				How: donor.CertificateProvider.Relationship,
			},
			Options: lpadata.CertificateProviderRelationshipValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.How.IsProfessionally() && donor.CertificateProvider.Relationship.IsPersonally() {
					donor.CertificateProvider.RelationshipLength = donordata.RelationshipLengthUnknown
				}

				if !donor.CertificateProvider.Relationship.Empty() && data.Form.How != donor.CertificateProvider.Relationship {
					donor.Tasks.CertificateProvider = task.StateInProgress
					donor.CertificateProvider.Address = place.Address{}
				}

				donor.CertificateProvider.Relationship = data.Form.How

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.CertificateProvider.Relationship.IsPersonally() {
					return page.Paths.HowLongHaveYouKnownCertificateProvider.Redirect(w, r, appData, donor)
				}

				return page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowYourCertificateProviderForm struct {
	How   lpadata.CertificateProviderRelationship
	Error error
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	how, err := lpadata.ParseCertificateProviderRelationship(page.PostFormString(r, "how"))

	return &howDoYouKnowYourCertificateProviderForm{
		How:   how,
		Error: err,
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.Error("how", "howYouKnowCertificateProvider", f.Error,
		validation.Selected())

	return errors
}
