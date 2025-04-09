package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatCountryDoYouLiveInData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *whatCountryDoYouLiveInForm
	Countries   []string
	CanTaskList bool
}

func WhatCountryDoYouLiveIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatCountryDoYouLiveInData{
			App: appData,
			Form: &whatCountryDoYouLiveInForm{
				CountryCode: provided.Donor.InternationalAddress.Country,
			},
			Countries:   place.Countries,
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatCountryDoYouLiveInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.CountryCode != provided.Donor.InternationalAddress.Country {
					provided.Donor.Address = place.Address{}
					if data.Form.CountryCode == "GB" {
						provided.Donor.InternationalAddress = place.InternationalAddress{}
					} else {
						provided.Donor.InternationalAddress = place.InternationalAddress{
							Country: data.Form.CountryCode,
						}
					}

					if err := donorStore.Put(r.Context(), provided); err != nil {
						return err
					}
				}

				if data.Form.CountryCode == "GB" {
					return donor.PathYourAddress.Redirect(w, r, appData, provided)
				} else {
					return donor.PathYourNonUKAddress.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}

type whatCountryDoYouLiveInForm struct {
	CountryCode string
}

func readWhatCountryDoYouLiveInForm(r *http.Request) *whatCountryDoYouLiveInForm {
	return &whatCountryDoYouLiveInForm{
		CountryCode: page.PostFormString(r, "country"),
	}
}

func (f *whatCountryDoYouLiveInForm) Validate() validation.List {
	var errors validation.List

	errors.Options("country", "countryYouLiveIn", []string{f.CountryCode},
		validation.Select(place.Countries...))

	return errors
}
