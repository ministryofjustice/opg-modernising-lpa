package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func ChooseAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorney, found := provided.Attorneys.Get(actoruid.FromRequest(r))
		if found == false {
			return donor.PathChooseAttorneys.Redirect(w, r, appData, provided)
		}

		data := newChooseAddressData(
			appData,
			"attorney",
			attorney.FullName(),
			attorney.UID,
		)

		if attorney.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &attorney.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				attorney.Address = address
				provided.Attorneys.Put(attorney)
				provided.UpdateDecisions()
				provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChooseAttorneysSummary.Redirect(w, r, appData, provided)
			}

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					return setAddress(*data.Form.Address)
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
				data.Addresses = provided.ActorAddresses()

			case "reuse-select":
				if data.Errors.None() {
					return setAddress(*data.Form.Address)
				} else {
					data.Addresses = provided.ActorAddresses()
				}
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			} else if action == "" && len(provided.ActorAddresses()) == 0 {
				data.Form.Action = "postcode"
			}
		}

		return tmpl(w, data)
	}
}
