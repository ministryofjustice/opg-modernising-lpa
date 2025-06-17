package donorpage

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Form                     *chooseAttorneysForm
	Donor                    *donordata.Provided
	Attorneys                []donordata.Attorney
	ShowTrustCorporationLink bool
}

func ChooseAttorneys(tmpl template.Template, service AttorneyService, newUID func() actoruid.UID) Handler {
	enterPath := donor.PathEnterAttorney
	summaryPath := donor.PathChooseAttorneysSummary
	if service.IsReplacement() {
		enterPath = donor.PathEnterReplacementAttorney
		summaryPath = donor.PathChooseReplacementAttorneysSummary
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		attorneys, err := service.Reusable(r.Context(), provided)
		if err != nil {
			return err
		}
		if len(attorneys) == 0 {
			return enterPath.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
		}

		data := &chooseAttorneysData{
			App:                      appData,
			Form:                     &chooseAttorneysForm{},
			Donor:                    provided,
			Attorneys:                attorneys,
			ShowTrustCorporationLink: provided.CanAddTrustCorporation(),
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)

			if data.Errors.None() {
				if len(data.Form.Indices) == 0 {
					return enterPath.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
				}

				var chosen []donordata.Attorney
				for _, index := range data.Form.Indices {
					chosen = append(chosen, attorneys[index])
				}

				if err := service.PutMany(r.Context(), provided, chosen); err != nil {
					return err
				}

				return summaryPath.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysForm struct {
	Indices []int
}

func readChooseAttorneysForm(r *http.Request) *chooseAttorneysForm {
	r.ParseForm()

	var indices []int
	for _, v := range r.PostForm["option"] {
		if index, err := strconv.Atoi(v); err == nil {
			indices = append(indices, index)
		}
	}

	return &chooseAttorneysForm{
		Indices: indices,
	}
}
