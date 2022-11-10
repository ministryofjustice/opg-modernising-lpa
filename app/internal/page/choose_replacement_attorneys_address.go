package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type chooseReplacementAttorneysAddressData struct {
	App       AppData
	Errors    map[string]string
	Attorney  Attorney
	Addresses []place.Address
	Form      *chooseAttorneysAddressForm
}

func ChooseReplacementAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		attorneyId := r.FormValue("id")
		ra, _ := lpa.GetReplacementAttorney(attorneyId)

		data := &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form:     &chooseAttorneysAddressForm{},
		}

		if ra.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &ra.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.Action == "manual" && len(data.Errors) == 0 {
				ra.Address = *data.Form.Address
				lpa.PutReplacementAttorney(ra)

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = chooseReplacementAttorneysSummaryPath
				}

				appData.Lang.Redirect(w, r, from, http.StatusFound)
				return nil
			}

			// Force the manual address view after selecting
			if data.Form.Action == "select" && len(data.Errors) == 0 {
				data.Form.Action = "manual"

				ra.Address = *data.Form.Address
				lpa.PutReplacementAttorney(ra)

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
			}

			if data.Form.Action == "lookup" && len(data.Errors) == 0 ||
				data.Form.Action == "select" && len(data.Errors) > 0 {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)
					data.Errors["lookup-postcode"] = "couldNotLookupPostcode"
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
