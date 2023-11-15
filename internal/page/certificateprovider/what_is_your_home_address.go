package certificateprovider

import (
	"context"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterHomeAddressData struct {
	App       page.AppData
	Addresses []place.Address
	Form      *form.AddressForm
	Errors    validation.List
}

func EnterHomeAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &enterHomeAddressData{
			App:  appData,
			Form: &form.AddressForm{},
		}

		if certificateProvider.HomeAddress.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &certificateProvider.HomeAddress
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

					if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
						return err
					}

					return appData.Redirect(w, r, lpa, page.Paths.CertificateProvider.ConfirmYourDetails.Format(certificateProvider.LpaID))
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress(r.Context(), logger, addressClient, data)
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress(r.Context(), logger, addressClient, data)
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
			} else {
				data.Form.Action = "postcode"
			}
		}

		return tmpl(w, data)
	}
}

func lookupAddress(ctx context.Context, logger Logger, addressClient AddressClient, data *enterHomeAddressData) {
	addresses, err := addressClient.LookupPostcode(ctx, data.Form.LookupPostcode)
	if err != nil {
		logger.Print(err)

		if errors.As(err, &place.BadRequestError{}) {
			data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
		} else {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
		}

		data.Form.Action = "postcode"
	} else if len(addresses) == 0 {
		data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noYourHomeAddressesFound"})
		data.Form.Action = "postcode"
	}

	data.Addresses = addresses
}
