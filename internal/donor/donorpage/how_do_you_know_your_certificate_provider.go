package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
	Options             donordata.CertificateProviderRelationshipOptions
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				How: donor.CertificateProvider.Relationship,
			},
			Options: donordata.CertificateProviderRelationshipValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.How.IsProfessionally() && donor.CertificateProvider.Relationship.IsPersonally() {
					donor.CertificateProvider.RelationshipLength = donordata.RelationshipLengthUnknown
				}

				if !donor.CertificateProvider.Relationship.Empty() && data.Form.How != donor.CertificateProvider.Relationship {
					donor.Tasks.CertificateProvider = actor.TaskInProgress
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
	How   donordata.CertificateProviderRelationship
	Error error
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	how, err := donordata.ParseCertificateProviderRelationship(page.PostFormString(r, "how"))

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
