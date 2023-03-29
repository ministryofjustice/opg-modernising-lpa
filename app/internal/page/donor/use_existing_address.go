package donor

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type useExistingAddressData struct {
	App       page.AppData
	Errors    validation.List
	Addresses []page.AddressDetail
	Subject   actor.Attorney
	Form      *UseExistingAddressForm
}

func UseExistingAddress(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneyType := r.FormValue("type")
		subjectId := r.FormValue("subjectId")

		subject, found := getSubject(attorneyType, subjectId, lpa)

		if !found {
			return fmt.Errorf("%s not found", attorneyType)
		}

		addresses := lpa.ActorAddresses()

		if len(addresses) == 0 || (len(addresses) == 1 && addressDetailsContains(subject, addresses)) {
			return appData.Redirect(w, r, lpa, r.FormValue("from"))
		}

		data := useExistingAddressData{
			App:       appData,
			Addresses: addresses,
			Subject:   subject,
		}

		if r.Method == http.MethodPost {
			data.Form = readUseExistingAddressForm(r)
			var options []string

			for i := 0; i < len(addresses); i++ {
				options = append(options, strconv.Itoa(i))
			}

			data.Errors = data.Form.Validate(options)

			if data.Errors.None() {
				addressIndex, _ := strconv.Atoi(data.Form.AddressIndex)
				subject.Address = addresses[addressIndex].Address

				redirect := appData.Paths.ChooseAttorneysSummary

				if attorneyType == "attorney" {
					lpa.Attorneys.Put(subject)
					lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
					lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)
				} else {
					lpa.ReplacementAttorneys.Put(subject)
					lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)
					redirect = appData.Paths.ChooseReplacementAttorneysSummary
				}

				err = lpaStore.Put(r.Context(), lpa)
				if err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirect)
			}
		}

		return tmpl(w, data)
	}
}

type UseExistingAddressForm struct {
	AddressIndex string
}

func readUseExistingAddressForm(r *http.Request) *UseExistingAddressForm {
	return &UseExistingAddressForm{
		AddressIndex: page.PostFormString(r, "address-index"),
	}
}

func (f UseExistingAddressForm) Validate(options []string) validation.List {
	errors := validation.List{}

	errors.String("address-index", "address", f.AddressIndex,
		validation.Select(options...))

	return errors
}

func getSubject(attorneyType, id string, lpa *page.Lpa) (actor.Attorney, bool) {
	if attorneyType == "attorney" {
		return lpa.Attorneys.Get(id)
	} else {
		return lpa.ReplacementAttorneys.Get(id)
	}
}

func addressDetailsContains(attorney actor.Attorney, addressDetails []page.AddressDetail) bool {
	for _, ad := range addressDetails {
		if ad.ID == attorney.ID && ad.Name == fmt.Sprintf("%s %s", attorney.FirstNames, attorney.LastName) {
			return true
		}
	}

	return false
}
