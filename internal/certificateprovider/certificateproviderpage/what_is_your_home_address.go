package certificateproviderpage

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatIsYourHomeAddressData struct {
	App       appcontext.Data
	Addresses []place.Address
	Form      *form.AddressForm
	Errors    validation.List
}

func WhatIsYourHomeAddress(logger Logger, tmpl template.Template, addressClient AddressClient, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		f := form.NewAddressForm()
		f.Action = "postcode"

		data := &whatIsYourHomeAddressData{
			App:  appData,
			Form: f,
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
					certificateProvider.HomeAddress = *data.Form.Address

					if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
						return err
					}

					return certificateprovider.PathYourPreferredLanguage.Redirect(w, r, appData, certificateProvider.LpaID)
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
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}

func lookupAddress(ctx context.Context, logger Logger, addressClient AddressClient, data *whatIsYourHomeAddressData) {
	addresses, err := addressClient.LookupPostcode(ctx, data.Form.LookupPostcode)
	if err != nil {
		logger.InfoContext(ctx, "postcode lookup", slog.Any("err", err))

		if errors.As(err, &place.BadRequestError{}) {
			data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
		} else {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
		}

		data.Form.Action = "postcode"
	} else if len(addresses) == 0 {
		data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"})
		data.Form.Action = "postcode"
	}

	data.Addresses = addresses
}
