package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func ChooseAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		attorney, found := donor.Attorneys.Get(actor.UIDFromRequest(r))

		if found == false {
			return page.Paths.ChooseAttorneys.Redirect(w, r, appData, donor)
		}

		data := newChooseAddressData(
			appData,
			"attorney",
			attorney.FullName(),
			attorney.UID,
			true,
		)

		data.ActorLabel = "attorney"
		data.FullName = attorney.FullName()
		data.UID = attorney.UID
		data.CanSkip = true

		if attorney.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &attorney.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				attorney.Address = address
				donor.Attorneys.Put(attorney)
				donor.Tasks.ChooseAttorneys = page.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				return donorStore.Put(r.Context(), donor)
			}

			switch data.Form.Action {
			case "skip":
				if err := setAddress(place.Address{}); err != nil {
					return err
				}

				return page.Paths.ChooseAttorneysSummary.Redirect(w, r, appData, donor)

			case "manual":
				if data.Errors.None() {
					if err := setAddress(*data.Form.Address); err != nil {
						return err
					}

					return page.Paths.ChooseAttorneysSummary.Redirect(w, r, appData, donor)
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
				data.Addresses = donor.ActorAddresses()

			case "reuse-select":
				if data.Errors.None() {
					if err := setAddress(*data.Form.Address); err != nil {
						return err
					}

					return page.Paths.ChooseAttorneysSummary.Redirect(w, r, appData, donor)
				} else {
					data.Addresses = donor.ActorAddresses()
				}
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
