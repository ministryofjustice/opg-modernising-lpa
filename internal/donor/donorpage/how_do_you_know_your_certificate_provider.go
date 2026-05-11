package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *newforms.EnumForm[lpadata.CertificateProviderRelationship, lpadata.CertificateProviderRelationshipOptions, *lpadata.CertificateProviderRelationship]
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: provided.CertificateProvider,
			Form:                newforms.NewEnumForm[lpadata.CertificateProviderRelationship](appData.Localizer.T("howYouKnowCertificateProvider"), lpadata.CertificateProviderRelationshipValues),
		}

		data.Form.Enum.SetInput(provided.CertificateProvider.Relationship)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if data.Form.Enum.Value.IsProfessionally() && provided.CertificateProvider.Relationship.IsPersonally() {
				provided.CertificateProvider.RelationshipLength = donordata.RelationshipLengthUnknown
			}

			if !provided.CertificateProvider.Relationship.Empty() && data.Form.Enum.Value != provided.CertificateProvider.Relationship {
				provided.Tasks.CertificateProvider = task.StateInProgress
				provided.CertificateProvider.Address = place.Address{}
			}

			provided.CertificateProvider.Relationship = data.Form.Enum.Value

			if err := reuseStore.PutCertificateProvider(r.Context(), provided.CertificateProvider); err != nil {
				return fmt.Errorf("put certificate provider reuse data: %w", err)
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if provided.CertificateProvider.Relationship.IsPersonally() {
				return donor.PathHowLongHaveYouKnownCertificateProvider.Redirect(w, r, appData, provided)
			}

			return donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
