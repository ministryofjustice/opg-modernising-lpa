package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type removeAttorneyData struct {
	App        appcontext.Data
	TitleLabel string
	Name       string
	Errors     validation.List
	Form       *form.YesNoForm
}

func RemoveAttorney(tmpl template.Template, service AttorneyService) Handler {
	titleLabel := "removeAnAttorney"
	summaryPath := donor.PathChooseAttorneysSummary
	errorLabel := "yesToRemoveAttorney"
	if service.IsReplacement() {
		titleLabel = "doYouWantToRemoveReplacementAttorney"
		summaryPath = donor.PathChooseReplacementAttorneysSummary
		errorLabel = "yesToRemoveReplacementAttorney"
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorneys := provided.Attorneys
		if service.IsReplacement() {
			attorneys = provided.ReplacementAttorneys
		}

		attorney, found := attorneys.Get(actoruid.FromRequest(r))
		if found == false {
			return summaryPath.Redirect(w, r, appData, provided)
		}

		data := &removeAttorneyData{
			App:        appData,
			TitleLabel: titleLabel,
			Name:       attorney.FullName(),
			Form:       form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, errorLabel)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					if err := service.Delete(r.Context(), provided, attorney); err != nil {
						return err
					}
				}

				return summaryPath.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
