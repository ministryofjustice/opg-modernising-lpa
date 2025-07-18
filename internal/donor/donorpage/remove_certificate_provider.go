package donorpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeCertificateProviderData struct {
	App    appcontext.Data
	Name   string
	Errors validation.List
	Form   *form.YesNoForm
}

func RemoveCertificateProvider(tmpl template.Template, service CertificateProviderService, notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &removeCertificateProviderData{
			App:  appData,
			Name: provided.CertificateProvider.FullName(),
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveCertificateProvider")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					if !provided.CertificateProviderInvitedAt.IsZero() {
						if err := notifyClient.SendActorEmail(r.Context(), notify.ToCertificateProvider(provided.CertificateProvider), provided.LpaUID, notify.CertificateProviderRemoved{
							DonorFullName:                  provided.Donor.FullName(),
							CertificateProviderFullName:    provided.CertificateProvider.FullName(),
							CertificateProviderInvitedDate: appData.Localizer.FormatDate(provided.CertificateProviderInvitedAt),
							LpaType:                        appData.Localizer.T(provided.Type.String()),
							LpaUID:                         provided.LpaUID,
							CertificateProviderStartURL:    appPublicURL + page.PathCertificateProviderStart.Format(),
						}); err != nil {
							return fmt.Errorf("send certificate provider removed email: %w", err)
						}
					}

					if err := service.Delete(r.Context(), provided); err != nil {
						return err
					}

					return donor.PathChooseCertificateProvider.RedirectQuery(w, r, appData, provided, url.Values{"removed": {data.Name}})
				}

				return donor.PathCertificateProviderSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
