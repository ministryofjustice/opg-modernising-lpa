package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func CertificateProviderAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := newChooseAddressData(
			appData,
			"certificateProvider",
			lpa.CertificateProvider.FullName(),
			"",
			false,
		)

		// so keys are set when amending address
		if lpa.CertificateProvider.Relationship.IsProfessionally() {
			data.overrideTitleKeys(titleKeys{
				Manual:                          "personsWorkAddress",
				PostcodeSelectAndPostcodeLookup: "selectPersonsWorkAddress",
				Postcode:                        "whatIsPersonsWorkPostcode",
				ReuseAndReuseSelect:             "selectAnAddressForPerson",
				ReuseOrNew:                      "addPersonsAddress",
			})
		}

		if lpa.CertificateProvider.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.CertificateProvider.Address
		} else if lpa.CertificateProvider.Relationship.IsProfessionally() {
			data.Form.Action = "postcode"
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				lpa.CertificateProvider.Address = *data.Form.Address
				lpa.Tasks.CertificateProvider = actor.TaskCompleted

				return donorStore.Put(r.Context(), lpa)
			}

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					if err := setAddress(*data.Form.Address); err != nil {
						return err
					}

					return page.Paths.TaskList.Redirect(w, r, appData, lpa)
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress(r.Context(), logger, addressClient, data, false)
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress(r.Context(), logger, addressClient, data, false)
				} else {
					data.Form.Action = "postcode"
				}

			case "reuse":
				data.Addresses = lpa.ActorAddresses()

			case "reuse-select":
				if data.Errors.None() {
					if err := setAddress(*data.Form.Address); err != nil {
						return err
					}

					return page.Paths.TaskList.Redirect(w, r, appData, lpa)
				} else {
					data.Addresses = lpa.ActorAddresses()
				}
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue("action")
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
