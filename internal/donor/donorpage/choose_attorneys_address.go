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

func ChooseAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, service AttorneyService) Handler {
	actorLabel := "attorney"
	choosePath := donor.PathChooseAttorneys
	summaryPath := donor.PathChooseAttorneysSummary
	if service.IsReplacement() {
		actorLabel = "replacementAttorney"
		choosePath = donor.PathChooseReplacementAttorneys
		summaryPath = donor.PathChooseReplacementAttorneysSummary
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorneys := provided.Attorneys
		if service.IsReplacement() {
			attorneys = provided.ReplacementAttorneys
		}

		attorney, found := attorneys.Get(actoruid.FromRequest(r))
		if found == false {
			return choosePath.Redirect(w, r, appData, provided)
		}

		data := newChooseAddressData(
			appData,
			actorLabel,
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

				if err := service.Put(r.Context(), provided, attorney); err != nil {
					return err
				}

				return summaryPath.Redirect(w, r, appData, provided)
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
