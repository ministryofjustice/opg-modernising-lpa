package donor

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

func newChooseAddressData(appData page.AppData, actorLabel, fullName string, UID actoruid.UID) *chooseAddressData {
	return &chooseAddressData{
		App:        appData,
		ActorLabel: actorLabel,
		FullName:   fullName,
		UID:        UID,
		Form:       form.NewAddressForm(),
		TitleKeys: titleKeys{
			Manual:                          "personsAddress",
			Postcode:                        "whatIsPersonsPostcode",
			PostcodeSelectAndPostcodeLookup: "selectAnAddressForPerson",
			ReuseAndReuseSelect:             "selectAnAddressForPerson",
			ReuseOrNew:                      "addPersonsAddress",
		},
	}
}

type chooseAddressData struct {
	App              page.AppData
	Errors           validation.List
	ActorLabel       string
	FullName         string
	UID              actoruid.UID
	Addresses        []place.Address
	Form             *form.AddressForm
	TitleKeys        titleKeys
	MakingAnotherLPA bool
	CanTaskList      bool
}

type titleKeys struct {
	Manual                          string
	PostcodeSelectAndPostcodeLookup string
	Postcode                        string
	ReuseAndReuseSelect             string
	ReuseOrNew                      string
}

func (d *chooseAddressData) overrideTitleKeys(newTitleKeys titleKeys) {
	d.TitleKeys = newTitleKeys
}

func lookupAddress(ctx context.Context, logger Logger, addressClient AddressClient, data *chooseAddressData, your bool) {
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
		if your {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"})
		} else {
			data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noAddressesFound"})
		}

		data.Form.Action = "postcode"
	}

	data.Addresses = addresses
}
