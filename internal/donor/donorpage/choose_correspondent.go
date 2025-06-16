package donorpage

import (
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseCorrespondentData struct {
	App            appcontext.Data
	Errors         validation.List
	Form           *chooseCorrespondentForm
	Donor          *donordata.Provided
	Correspondents []donordata.Correspondent
}

func ChooseCorrespondent(tmpl template.Template, service CorrespondentService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		correspondents, err := service.Reusable(r.Context())
		if err != nil {
			return err
		}
		if len(correspondents) == 0 {
			return donor.PathEnterCorrespondentDetails.Redirect(w, r, appData, provided)
		}

		data := &chooseCorrespondentData{
			App:            appData,
			Form:           &chooseCorrespondentForm{},
			Donor:          provided,
			Correspondents: correspondents,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseCorrespondentForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.New {
					return donor.PathEnterCorrespondentDetails.Redirect(w, r, appData, provided)
				}

				provided.Correspondent = correspondents[data.Form.Index]

				if err := service.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathCorrespondentSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type chooseCorrespondentForm struct {
	New   bool
	Index int
	Err   error
}

func readChooseCorrespondentForm(r *http.Request) *chooseCorrespondentForm {
	option := page.PostFormString(r, "option")
	index, err := strconv.Atoi(option)

	return &chooseCorrespondentForm{
		New:   option == "new",
		Index: index,
		Err:   err,
	}
}

func (f *chooseCorrespondentForm) Validate() validation.List {
	var errors validation.List

	if !f.New && f.Err != nil {
		errors.Add("option", validation.SelectError{Label: "aCorrespondentOrToAddANewCorrespondent"})
	}

	return errors
}
