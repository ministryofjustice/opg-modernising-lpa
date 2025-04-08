package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourNonUKAddressData struct {
	App              appcontext.Data
	Errors           validation.List
	Form             *yourNonUKAddressForm
	Country          string
	CanTaskList      bool
	MakingAnotherLPA bool
}

func YourNonUKAddress(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourNonUKAddressData{
			App: appData,
			Form: &yourNonUKAddressForm{
				Address: provided.Donor.InternationalAddress,
			},
			Country:          provided.Donor.InternationalAddress.Country,
			CanTaskList:      !provided.Type.Empty(),
			MakingAnotherLPA: r.FormValue("makingAnotherLPA") == "1",
		}

		if r.Method == http.MethodPost {
			data.Form = readYourNonUKAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				data.Form.Address.Country = data.Country
				addressChangesMade := provided.Donor.InternationalAddress != data.Form.Address

				if addressChangesMade {
					provided.HasSentApplicationUpdatedEvent = false
					provided.Donor.InternationalAddress = data.Form.Address
					provided.Donor.Address = data.Form.Address.ToAddress()

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}
				}

				if data.MakingAnotherLPA {
					if !addressChangesMade {
						return donor.PathMakeANewLPA.Redirect(w, r, appData, provided)
					}

					return donor.PathWeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, provided, url.Values{"detail": {"address"}})
				}

				if appData.SupporterData != nil {
					return donor.PathYourEmail.Redirect(w, r, appData, provided)
				}

				return donor.PathReceivingUpdatesAboutYourLpa.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourNonUKAddressForm struct {
	Address place.InternationalAddress
}

func readYourNonUKAddressForm(r *http.Request) *yourNonUKAddressForm {
	return &yourNonUKAddressForm{
		Address: place.InternationalAddress{
			ApartmentNumber: page.PostFormString(r, "apartmentNumber"),
			BuildingNumber:  page.PostFormString(r, "buildingNumber"),
			BuildingName:    page.PostFormString(r, "buildingName"),
			StreetName:      page.PostFormString(r, "streetName"),
			Town:            page.PostFormString(r, "town"),
			Region:          page.PostFormString(r, "region"),
			PostalCode:      page.PostFormString(r, "postalCode"),
		},
	}
}

func (f *yourNonUKAddressForm) Validate() validation.List {
	var errors validation.List

	if f.Address.ApartmentNumber == "" && f.Address.BuildingNumber == "" && f.Address.BuildingName == "" {
		errors.Add("buildingAddress", validation.EnterError{Label: "atLeastOneBuildingAddress"})
	}

	errors.String("town", "townSuburbOrCity", f.Address.Town,
		validation.Empty())

	return errors
}
