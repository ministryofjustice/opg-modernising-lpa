package donor

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
			"",
			false,
		)

		makingAnotherLPA := r.URL.Query().Get("from") == page.Paths.MakeANewLPA.Format(donor.LpaID)

		if makingAnotherLPA {
			previousDetails, err := donorStore.Latest(r.Context())
			if err != nil {
				return err
			}

			data.Form.Action = "manual"
			data.Form.Address = &previousDetails.Donor.Address
		}

		if donor.Donor.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &donor.Donor.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(true)

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					if donor.Donor.Address.Postcode != data.Form.Address.Postcode {
						donor.HasSentApplicationUpdatedEvent = false
					}

					donor.Donor.Address = *data.Form.Address

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}

					if makingAnotherLPA {
						r.Form.Del("from")
						return page.Paths.WeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, donor, url.Values{"detail": {"address"}})
					}

					return page.Paths.YourPreferredLanguage.Redirect(w, r, appData, donor)
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
			action := r.FormValue("action")
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
