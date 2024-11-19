package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *form.SelectForm[lpadata.CertificateProviderRelationship, lpadata.CertificateProviderRelationshipOptions, *lpadata.CertificateProviderRelationship]
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: provided.CertificateProvider,
			Form:                form.NewSelectForm(provided.CertificateProvider.Relationship, lpadata.CertificateProviderRelationshipValues, "howYouKnowCertificateProvider"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.Selected.IsProfessionally() && provided.CertificateProvider.Relationship.IsPersonally() {
					provided.CertificateProvider.RelationshipLength = donordata.RelationshipLengthUnknown
				}

				if !provided.CertificateProvider.Relationship.Empty() && data.Form.Selected != provided.CertificateProvider.Relationship {
					provided.Tasks.CertificateProvider = task.StateInProgress
					provided.CertificateProvider.Address = place.Address{}
				}

				provided.CertificateProvider.Relationship = data.Form.Selected

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.CertificateProvider.Relationship.IsPersonally() {
					return donor.PathHowLongHaveYouKnownCertificateProvider.Redirect(w, r, appData, provided)
				}

				return donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
