package certificateproviderpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeCertificateProviderData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmDontWantToBeCertificateProvider(tmpl template.Template, lpaStoreClient LpaStoreClient, donorStore DonorStore, certificateProviderStore CertificateProviderStore, notifyClient NotifyClient, donorStartURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		data := &confirmDontWantToBeCertificateProviderData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			var email notify.Email

			if lpa.SignedForDonor() {
				email = notify.CertificateProviderOptedOutPostWitnessingEmail{
					Greeting:                      notifyClient.EmailGreeting(lpa),
					CertificateProviderFirstNames: lpa.CertificateProvider.FirstNames,
					CertificateProviderFullName:   lpa.CertificateProvider.FullName(),
					DonorFullName:                 lpa.Donor.FullName(),
					LpaType:                       appData.Localizer.T(lpa.Type.String()),
					LpaReferenceNumber:            lpa.LpaUID,
					DonorStartPageURL:             donorStartURL,
				}

				if !lpa.Status.IsCannotRegister() {
					if err := lpaStoreClient.SendCertificateProviderOptOut(r.Context(), lpa.LpaUID, lpa.CertificateProvider.UID); err != nil {
						return err
					}
				}
			} else {
				donor, err := donorStore.GetAny(r.Context())
				if err != nil {
					return err
				}

				email = notify.CertificateProviderOptedOutPreWitnessingEmail{
					Greeting:                    notifyClient.EmailGreeting(lpa),
					CertificateProviderFullName: donor.CertificateProvider.FullName(),
					DonorFullName:               donor.Donor.FullName(),
					LpaType:                     appData.Localizer.T(donor.Type.String()),
					LpaReferenceNumber:          donor.LpaUID,
					DonorStartPageURL:           donorStartURL,
				}

				donor.CertificateProvider = donordata.CertificateProvider{}
				donor.Tasks.CertificateProvider = task.StateNotStarted
				donor.Tasks.CheckYourLpa = task.StateNotStarted

				if err = donorStore.Put(r.Context(), donor); err != nil {
					return err
				}
			}

			if err := certificateProviderStore.Delete(r.Context()); err != nil {
				return err
			}

			if err := notifyClient.SendActorEmail(r.Context(), notify.ToLpaDonor(lpa), lpa.LpaUID, email); err != nil {
				return err
			}

			return page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider.RedirectQuery(w, r, appData, url.Values{
				"donorFullName":   {lpa.Donor.FullName()},
				"donorFirstNames": {lpa.Donor.FirstNames},
			})
		}

		return tmpl(w, data)
	}
}
