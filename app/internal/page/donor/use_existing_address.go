package donor

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type UseExistingAddressData struct {
	App       page.AppData
	Errors    validation.List
	Addresses []page.AddressDetail
	Subject   actor.Attorney
}

func UseExistingAddress(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneyType := r.FormValue("role")
		actorId := r.FormValue("id")

		subject, err := getSubject(attorneyType, actorId, lpa)
		if err != nil {
			return err
		}

		addresses := lpa.ActorAddresses()

		if len(addresses) == 0 || (len(addresses) == 1 && addressDetailsContains(subject, addresses)) {
			return appData.Redirect(w, r, lpa, r.FormValue("from"))
		}

		data := UseExistingAddressData{
			App:       appData,
			Addresses: addresses,
			Subject:   subject,
		}

		if r.Method == http.MethodPost {
			form := readUseExistingAddressForm(r)
			addressIndex, err := strconv.Atoi(form.Address)
			if err != nil {
				return err
			}

			subject.Address = addresses[addressIndex].Address

			redirect := appData.Paths.ChooseAttorneysSummary

			if attorneyType == "attorney" {
				ok := lpa.Attorneys.Put(subject)

				if !ok {
					return errors.New("attorney not found")
				}
			} else {
				ok := lpa.ReplacementAttorneys.Put(subject)

				if !ok {
					return errors.New("replacement attorney not found")
				}
				redirect = appData.Paths.ChooseReplacementAttorneysSummary
			}

			err = lpaStore.Put(r.Context(), lpa)
			if err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, redirect)
		}

		return tmpl(w, data)
	}
}

type UseExistingAddressForm struct {
	Address string
}

func readUseExistingAddressForm(r *http.Request) *UseExistingAddressForm {
	return &UseExistingAddressForm{
		Address: page.PostFormString(r, "address"),
	}
}

func getSubject(attorneyType, id string, lpa *page.Lpa) (actor.Attorney, error) {
	if attorneyType == "attorney" {
		attorney, found := lpa.Attorneys.Get(id)
		if !found {
			return actor.Attorney{}, errors.New("attorney not found")
		}

		return attorney, nil
	} else {
		replacementAttorney, found := lpa.ReplacementAttorneys.Get(id)
		if !found {
			return actor.Attorney{}, errors.New("replacement attorney not found")
		}

		return replacementAttorney, nil
	}
}

func addressDetailsContains(attorney actor.Attorney, addressDetails []page.AddressDetail) bool {
	for _, ad := range addressDetails {
		if ad.ID == attorney.ID {
			return true
		}
	}

	return false
}
