package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func YourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &chooseAddressData{
			App:  appData,
			Form: &form.AddressForm{},
		}

		if lpa.Donor.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.Donor.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(true)

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					if lpa.Donor.Address.Postcode != data.Form.Address.Postcode {
						lpa.HasSentApplicationUpdatedEvent = false
					}

					lpa.Donor.Address = *data.Form.Address

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}

					return appData.Redirect(w, r, lpa, page.Paths.WhoIsTheLpaFor.Format(lpa.ID))
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
