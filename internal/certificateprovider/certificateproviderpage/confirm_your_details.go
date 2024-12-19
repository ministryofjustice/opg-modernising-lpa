package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourDetailsData struct {
	App                    appcontext.Data
	Errors                 validation.List
	Lpa                    *lpadata.Lpa
	CertificateProvider    *certificateproviderdata.Provided
	PhoneNumberLabel       string
	AddressLabel           string
	DetailComponentContent string
	ShowPhone              bool
	ShowHomeAddress        bool
}

func ConfirmYourDetails(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ConfirmYourDetails = task.StateCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			redirect := certificateprovider.PathYourRole
			if lpa.SignedForDonor() {
				redirect = certificateprovider.PathTaskList
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
			ShowPhone:              lpa.CertificateProvider.Phone != "",
			ShowHomeAddress:        certificateProvider.HomeAddress.Line1 != "",
		}

		if !data.ShowPhone {
			data.DetailComponentContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentLayMissingPhone"
		}

		if lpa.Donor.Channel.IsPaper() {
			data.PhoneNumberLabel = "contactNumber"
		} else if lpa.CertificateProvider.Relationship.IsProfessionally() {
			data.AddressLabel = "workAddress"
			data.DetailComponentContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessional"

			if !data.ShowPhone {
				data.DetailComponentContent = "whatToDoIfAnyDetailsAreIncorrectCertificateProviderContentProfessionalMissingPhone"
			}
		}

		return tmpl(w, data)
	}
}
