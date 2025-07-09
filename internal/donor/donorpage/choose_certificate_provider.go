package donorpage

import (
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseCertificateProviderData struct {
	App                  appcontext.Data
	Errors               validation.List
	Form                 *chooseCertificateProviderForm
	Donor                *donordata.Provided
	CertificateProviders []donordata.CertificateProvider
}

func ChooseCertificateProvider(tmpl template.Template, service CertificateProviderService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		certificateProviders, err := service.Reusable(r.Context())
		if err != nil {
			return err
		}
		if len(certificateProviders) == 0 {
			return donor.PathCertificateProviderDetails.RedirectQuery(w, r, appData, provided, r.URL.Query())
		}

		data := &chooseCertificateProviderData{
			App:                  appData,
			Form:                 &chooseCertificateProviderForm{},
			Donor:                provided,
			CertificateProviders: certificateProviders,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.New {
					return donor.PathCertificateProviderDetails.Redirect(w, r, appData, provided)
				}

				provided.CertificateProvider = certificateProviders[data.Form.Index]
				if err := service.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathCertificateProviderSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type chooseCertificateProviderForm struct {
	New   bool
	Index int
	Err   error
}

func readChooseCertificateProviderForm(r *http.Request) *chooseCertificateProviderForm {
	option := page.PostFormString(r, "option")
	index, err := strconv.Atoi(option)

	return &chooseCertificateProviderForm{
		New:   option == "new",
		Index: index,
		Err:   err,
	}
}

func (f *chooseCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	if !f.New && f.Err != nil {
		errors.Add("option", validation.SelectError{Label: "aCertificateProviderOrToAddANewCertificateProvider"})
	}

	return errors
}
