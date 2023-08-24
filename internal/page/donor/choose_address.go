package donor

import (
	"context"
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAddressData struct {
	App        page.AppData
	Errors     validation.List
	ActorLabel string
	FullName   string
	ID         string
	CanSkip    bool
	Addresses  []place.Address
	Form       *form.AddressForm
}

func lookupAddress(ctx context.Context, logger Logger, addressClient AddressClient, data *chooseAddressData, your bool) {
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
		if your {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"})
		} else {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noAddressesFound"})
		}

		data.Form.Action = "postcode"
	}

	data.Addresses = addresses
}
