package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howLongHaveYouKnownCertificateProviderData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	RelationshipLength  donordata.CertificateProviderRelationshipLength
	Options             donordata.CertificateProviderRelationshipLengthOptions
}

func HowLongHaveYouKnownCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			RelationshipLength:  donor.CertificateProvider.RelationshipLength,
			Options:             donordata.CertificateProviderRelationshipLengthValues,
		}

		if r.Method == http.MethodPost {
			form := readHowLongHaveYouKnownCertificateProviderForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				if form.RelationshipLength == donordata.LessThanTwoYears {
					return page.Paths.ChooseNewCertificateProvider.Redirect(w, r, appData, donor)
				}

				donor.CertificateProvider.RelationshipLength = form.RelationshipLength
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type howLongHaveYouKnownCertificateProviderForm struct {
	RelationshipLength donordata.CertificateProviderRelationshipLength
	Error              error
}

func readHowLongHaveYouKnownCertificateProviderForm(r *http.Request) *howLongHaveYouKnownCertificateProviderForm {
	relationshipLength, err := donordata.ParseCertificateProviderRelationshipLength(page.PostFormString(r, "relationship-length"))

	return &howLongHaveYouKnownCertificateProviderForm{
		RelationshipLength: relationshipLength,
		Error:              err,
	}
}

func (f *howLongHaveYouKnownCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.Error("relationship-length", "howLongYouHaveKnownCertificateProvider", f.Error,
		validation.Selected())

	return errors
}