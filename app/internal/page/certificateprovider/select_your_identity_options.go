package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type selectYourIdentityOptionsData struct {
	App    page.AppData
	Errors validation.List
	Form   *selectYourIdentityOptionsForm
	Page   int
}

func SelectYourIdentityOptions(tmpl template.Template, pageIndex int, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &selectYourIdentityOptionsData{
			App:  appData,
			Page: pageIndex,
			Form: &selectYourIdentityOptionsForm{
				Selected: certificateProvider.IdentityOption,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSelectYourIdentityOptionsForm(r)
			data.Errors = data.Form.Validate(pageIndex)

			if data.Form.None {
				switch pageIndex {
				case 0:
					return appData.Redirect(w, r, nil, page.Paths.CertificateProviderSelectYourIdentityOptions1)
				case 1:
					return appData.Redirect(w, r, nil, page.Paths.CertificateProviderSelectYourIdentityOptions2)
				default:
					// will go to vouching flow when that is built
					return appData.Redirect(w, r, nil, page.Paths.CertificateProviderStart)
				}
			}

			if data.Errors.None() {
				certificateProvider.IdentityOption = data.Form.Selected

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return appData.Redirect(w, r, nil, page.Paths.CertificateProviderYourChosenIdentityOptions)
			}
		}

		return tmpl(w, data)
	}
}

type selectYourIdentityOptionsForm struct {
	Selected identity.Option
	None     bool
}

func readSelectYourIdentityOptionsForm(r *http.Request) *selectYourIdentityOptionsForm {
	option := page.PostFormString(r, "option")

	return &selectYourIdentityOptionsForm{
		Selected: identity.ReadOption(option),
		None:     option == "none",
	}
}

func (f *selectYourIdentityOptionsForm) Validate(pageIndex int) validation.List {
	var errors validation.List

	if f.Selected == identity.UnknownOption && !f.None {
		if pageIndex == 0 {
			errors.Add("option", validation.SelectError{Label: "fromTheListedOptions"})
		} else {
			errors.Add("option", validation.SelectError{Label: "whichDocumentYouWillUse"})
		}
	}

	return errors
}
