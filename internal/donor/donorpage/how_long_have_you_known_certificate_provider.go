package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howLongHaveYouKnownCertificateProviderData struct {
	App                 appcontext.Data
	Errors              validation.List
	Form                *newforms.EnumForm[donordata.CertificateProviderRelationshipLength, donordata.CertificateProviderRelationshipLengthOptions, *donordata.CertificateProviderRelationshipLength]
	CertificateProvider donordata.CertificateProvider
}

func HowLongHaveYouKnownCertificateProvider(tmpl template.Template, donorStore DonorStore, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			Form:                newforms.NewEnumForm[donordata.CertificateProviderRelationshipLength](appData.Localizer.T("howLongYouHaveKnownCertificateProvider"), donordata.CertificateProviderRelationshipLengthValues),
			CertificateProvider: provided.CertificateProvider,
		}

		data.Form.Enum.SetInput(provided.CertificateProvider.RelationshipLength)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if data.Form.Enum.Value == donordata.LessThanTwoYears {
				return donor.PathChooseNewCertificateProvider.Redirect(w, r, appData, provided)
			}

			provided.CertificateProvider.RelationshipLength = data.Form.Enum.Value

			if err := reuseStore.PutCertificateProvider(r.Context(), provided.CertificateProvider); err != nil {
				return fmt.Errorf("put certificate provider reuse data: %w", err)
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
