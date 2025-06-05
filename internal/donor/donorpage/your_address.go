package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func YourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := newChooseAddressData(
			appData,
			"",
			"",
			provided.Donor.UID,
		)

		if provided.Donor.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &provided.Donor.Address
		}

		data.MakingAnotherLPA = r.FormValue("makingAnotherLPA") == "1"
		data.CanTaskList = !provided.Type.Empty()

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(true)

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					addressChangesMade := provided.Donor.Address.Line1 != data.Form.Address.Line1 ||
						provided.Donor.Address.Line2 != data.Form.Address.Line2 ||
						provided.Donor.Address.Line3 != data.Form.Address.Line3 ||
						provided.Donor.Address.TownOrCity != data.Form.Address.TownOrCity ||
						provided.Donor.Address.Postcode != data.Form.Address.Postcode

					if addressChangesMade {
						provided.HasSentApplicationUpdatedEvent = false
						provided.Donor.Address = *data.Form.Address
						if err := donorStore.Put(r.Context(), provided); err != nil {
							return err
						}
					}

					next := determineNext(appData, provided, addressChangesMade, data)

					if addressChangesMade && provided.CertificateProvider.Address.Line1 == provided.Donor.Address.Line1 &&
						provided.CertificateProvider.Address.Postcode == provided.Donor.Address.Postcode {
						return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
							"warningFrom": {appData.Page},
							"next":        {next.Path},
							"actor":       {actor.TypeDonor.String()},
						})
					}

					if next.QueryRedirect {
						return next.QueryFunc(w, r, appData, provided, next.QueryValues)
					}
					return next.RedirectFunc(w, r, appData, provided)
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress(r.Context(), logger, addressClient, data, true)
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress(r.Context(), logger, addressClient, data, true)
				} else {
					data.Form.Action = "postcode"
				}
			}
		}

		if r.Method == http.MethodGet && data.Form.Address == nil {
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}

type NavigationDecision struct {
	Path          string
	RedirectFunc  func(http.ResponseWriter, *http.Request, appcontext.Data, *donordata.Provided) error
	QueryRedirect bool
	QueryFunc     func(http.ResponseWriter, *http.Request, appcontext.Data, *donordata.Provided, url.Values) error
	QueryValues   url.Values
}

func determineNext(appData appcontext.Data, provided *donordata.Provided, addressChangesMade bool, data *chooseAddressData) NavigationDecision {
	if data.MakingAnotherLPA {
		if !addressChangesMade {
			return NavigationDecision{
				Path:         donor.PathMakeANewLPA.Format(provided.LpaID),
				RedirectFunc: donor.PathMakeANewLPA.Redirect,
			}
		}
		return NavigationDecision{
			Path:          donor.PathWeHaveUpdatedYourDetails.Format(provided.LpaID),
			QueryRedirect: true,
			QueryFunc:     donor.PathWeHaveUpdatedYourDetails.RedirectQuery,
			QueryValues:   url.Values{"detail": {"address"}},
		}
	}

	if appData.SupporterData != nil {
		return NavigationDecision{
			Path:         donor.PathYourEmail.Format(provided.LpaID),
			RedirectFunc: donor.PathYourEmail.Redirect,
		}
	}

	return NavigationDecision{
		Path:         donor.PathReceivingUpdatesAboutYourLpa.Format(provided.LpaID),
		RedirectFunc: donor.PathReceivingUpdatesAboutYourLpa.Redirect,
	}
}
