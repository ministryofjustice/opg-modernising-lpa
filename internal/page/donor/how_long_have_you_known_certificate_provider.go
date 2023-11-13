package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howLongHaveYouKnownCertificateProviderData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	RelationshipLength  actor.CertificateProviderRelationshipLength
	Options             actor.CertificateProviderRelationshipLengthOptions
}

func HowLongHaveYouKnownCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			RelationshipLength:  lpa.CertificateProvider.RelationshipLength,
			Options:             actor.CertificateProviderRelationshipLengthValues,
		}

		if r.Method == http.MethodPost {
			form := readHowLongHaveYouKnownCertificateProviderForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				if form.RelationshipLength == actor.LessThanTwoYears {
					return appData.Redirect(w, r, lpa, page.Paths.ChooseNewCertificateProvider.Format(lpa.ID))
				}

				lpa.CertificateProvider.RelationshipLength = form.RelationshipLength
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type howLongHaveYouKnownCertificateProviderForm struct {
	RelationshipLength actor.CertificateProviderRelationshipLength
	Error              error
}

func readHowLongHaveYouKnownCertificateProviderForm(r *http.Request) *howLongHaveYouKnownCertificateProviderForm {
	relationshipLength, err := actor.ParseCertificateProviderRelationshipLength(page.PostFormString(r, "relationship-length"))

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
