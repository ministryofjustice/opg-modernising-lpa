package donorpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func CertificateProviderAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := newChooseAddressData(
			appData,
			"certificateProvider",
			provided.CertificateProvider.FullName(),
			provided.CertificateProvider.UID,
		)

		// so keys are set when amending address
		if provided.CertificateProvider.Relationship.IsProfessionally() {
			data.overrideTitleKeys(titleKeys{
				Manual:                          "personsWorkAddress",
				PostcodeSelectAndPostcodeLookup: "selectPersonsWorkAddress",
				Postcode:                        "whatIsPersonsWorkPostcode",
				ReuseAndReuseSelect:             "selectAnAddressForPerson",
				ReuseOrNew:                      "addPersonsAddress",
			})
		}

		if provided.CertificateProvider.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &provided.CertificateProvider.Address
		} else if provided.CertificateProvider.Relationship.IsProfessionally() {
			data.Form.Action = "postcode"
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				addressHasChanged := provided.CertificateProvider.AddressHasChanged(address)
				provided.CertificateProvider.Address = address
				provided.Tasks.CertificateProvider = task.StateCompleted

				// Allow changing address for certificate provider on the page they
				// witness, without having to be notified.
				if !provided.SignedAt.IsZero() {
					provided.UpdateCheckedHash()
				}

				if err := reuseStore.PutCertificateProvider(r.Context(), provided.CertificateProvider); err != nil {
					return fmt.Errorf("put certificate provider reuse data: %w", err)
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if addressHasChanged && provided.CertificateProviderSharesAddress() {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {donor.PathCertificateProviderSummary.Format(provided.LpaID)},
						"actor":       {actor.TypeCertificateProvider.String()},
					})
				}

				return donor.PathCertificateProviderSummary.Redirect(w, r, appData, provided)
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
