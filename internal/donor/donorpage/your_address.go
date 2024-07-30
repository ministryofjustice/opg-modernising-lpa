package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func YourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := newChooseAddressData(
			appData,
			"",
			"",
			donor.Donor.UID,
		)

		if donor.Donor.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &donor.Donor.Address
		}

		data.MakingAnotherLPA = r.FormValue("makingAnotherLPA") == "1"
		data.CanTaskList = !donor.Type.Empty()

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(true)

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					addressChangesMade := donor.Donor.Address.Line1 != data.Form.Address.Line1 ||
						donor.Donor.Address.Line2 != data.Form.Address.Line2 ||
						donor.Donor.Address.Line3 != data.Form.Address.Line3 ||
						donor.Donor.Address.TownOrCity != data.Form.Address.TownOrCity ||
						donor.Donor.Address.Postcode != data.Form.Address.Postcode

					if addressChangesMade {
						donor.HasSentApplicationUpdatedEvent = false
						donor.Donor.Address = *data.Form.Address
						if err := donorStore.Put(r.Context(), donor); err != nil {
							return err
						}
					}

					if data.MakingAnotherLPA {
						if !addressChangesMade {
							return page.Paths.MakeANewLPA.Redirect(w, r, appData, donor)
						}

						return page.Paths.WeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, donor, url.Values{"detail": {"address"}})
					}

					if appData.SupporterData != nil {
						return page.Paths.YourEmail.Redirect(w, r, appData, donor)
					}

					return page.Paths.CanYouSignYourLpa.Redirect(w, r, appData, donor)
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress(r.Context(), logger, addressClient, data, true)
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress(r.Context(), logger, addressClient, data, true)
				} else {
					data.Form.Action = "postcode"
				}
			}
		}

		if r.Method == http.MethodGet && data.Form.Address == nil {
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
