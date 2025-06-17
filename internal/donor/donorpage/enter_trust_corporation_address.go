package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func EnterTrustCorporationAddress(logger Logger, tmpl template.Template, addressClient AddressClient, service AttorneyService) Handler {
	summaryPath := donor.PathChooseAttorneysSummary
	if service.IsReplacement() {
		summaryPath = donor.PathChooseReplacementAttorneysSummary
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporation := provided.Attorneys.TrustCorporation
		if service.IsReplacement() {
			trustCorporation = provided.ReplacementAttorneys.TrustCorporation
		}

		data := newChooseAddressData(
			appData,
			"theTrustCorporation",
			"",
			trustCorporation.UID,
		)

		if trustCorporation.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &trustCorporation.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				trustCorporation.Address = address

				if err := service.PutTrustCorporation(r.Context(), provided, trustCorporation); err != nil {
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
			}
		}

		return tmpl(w, data)
	}
}
