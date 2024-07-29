package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                    page.AppData
	Errors                 validation.List
	Lpa                    *lpastore.Lpa
	CertificateProvider    *certificateproviderdata.Provided
	PhoneNumberLabel       string
	AddressLabel           string
	DetailComponentContent string
}

func ConfirmYourDetails(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			redirect := page.Paths.CertificateProvider.YourRole
			if !lpa.SignedAt.IsZero() {
				redirect = page.Paths.CertificateProvider.TaskList
			}

			return redirect.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		data := &confirmYourDetailsData{
			App:                    appData,
			CertificateProvider:    certificateProvider,
			Lpa:                    lpa,
			PhoneNumberLabel:       "mobileNumber",
			AddressLabel:           "address",
			DetailComponentContent: "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLay",
		}

		if lpa.Donor.Channel.IsPaper() {
			data.PhoneNumberLabel = "contactNumber"
		}

		if lpa.CertificateProvider.Relationship.IsProfessionally() {
			data.AddressLabel = "workAddress"
			data.DetailComponentContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional"
		}

		return tmpl(w, data)
	}
}
