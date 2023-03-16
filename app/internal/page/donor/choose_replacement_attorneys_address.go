package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseReplacementAttorneysAddressData struct {
	App       page.AppData
	Errors    validation.List
	Attorney  actor.Attorney
	Addresses []place.Address
	Form      *form.AddressForm
}

func ChooseReplacementAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneyId := r.FormValue("id")
		ra, _ := lpa.ReplacementAttorneys.Get(attorneyId)

		data := &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form:     &form.AddressForm{},
		}

		if ra.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &ra.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			if data.Form.Action == "manual" && data.Errors.None() {
				ra.Address = *data.Form.Address
				lpa.ReplacementAttorneys.Put(ra)
				lpa.Tasks.ChooseReplacementAttorneys = page.TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = appData.Paths.ChooseReplacementAttorneysSummary
				}

				return appData.Redirect(w, r, lpa, from)
			}

			// Force the manual address view after selecting
			if data.Form.Action == "select" && data.Errors.None() {
				data.Form.Action = "manual"

				ra.Address = *data.Form.Address
				lpa.ReplacementAttorneys.Put(ra)

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}
			}

			if data.Form.Action == "lookup" && data.Errors.None() ||
				data.Form.Action == "select" && data.Errors.Any() {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)

					if errors.As(err, &place.BadRequestError{}) {
						data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
					} else {
						data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
					}
				} else if len(addresses) == 0 {
					data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noAddressesFound"})
				}

				data.Addresses = addresses
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
