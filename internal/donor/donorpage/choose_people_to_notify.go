package donorpage

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifyData struct {
	App                      appcontext.Data
	Errors                   validation.List
	Form                     *choosePeopleToNotifyForm
	Donor                    *donordata.Provided
	PeopleToNotify           []donordata.PersonToNotify
	ShowTrustCorporationLink bool
}

func ChoosePeopleToNotify(tmpl template.Template, service PeopleToNotifyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		peopleToNotify, err := service.Reusable(r.Context(), provided)
		if err != nil {
			return err
		}
		if len(peopleToNotify) == 0 {
			return donor.PathEnterPersonToNotify.RedirectQuery(w, r, appData, provided, url.Values{
				"addAnother": {r.FormValue("addAnother")},
			})
		}

		data := &choosePeopleToNotifyData{
			App:                      appData,
			Form:                     &choosePeopleToNotifyForm{},
			Donor:                    provided,
			PeopleToNotify:           peopleToNotify,
			ShowTrustCorporationLink: provided.CanAddTrustCorporation(),
		}

		if r.Method == http.MethodPost {
			data.Form = readChoosePeopleToNotifyForm(r)
			data.Errors = data.Form.Validate(len(provided.PeopleToNotify))

			if data.Errors.None() {
				if len(data.Form.Indices) == 0 {
					return donor.PathEnterPersonToNotify.RedirectQuery(w, r, appData, provided, url.Values{
						"addAnother": {r.FormValue("addAnother")},
					})
				}

				var people []donordata.PersonToNotify
				for _, index := range data.Form.Indices {
					people = append(people, peopleToNotify[index])
				}

				if err := service.PutMany(r.Context(), provided, people); err != nil {
					return err
				}

				return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type choosePeopleToNotifyForm struct {
	Indices []int
}

func readChoosePeopleToNotifyForm(r *http.Request) *choosePeopleToNotifyForm {
	r.ParseForm()

	var indices []int
	for _, v := range r.PostForm["option"] {
		if index, err := strconv.Atoi(v); err == nil {
			indices = append(indices, index)
		}
	}

	return &choosePeopleToNotifyForm{
		Indices: indices,
	}
}

func (f *choosePeopleToNotifyForm) Validate(currentCount int) (errors validation.List) {
	if len(f.Indices) > 5 {
		errors.Add("option", validation.CustomError{Label: "youCannotSelectMoreThanFivePeopleToNotify"})
	} else if len(f.Indices)+currentCount > 5 {
		errors.Add("option", validation.CustomError{Label: "yourSelectionTakesPeopleToNotifyOverFive"})
	}

	return errors
}

func (f *choosePeopleToNotifyForm) Selected(index int) bool {
	return slices.Contains(f.Indices, index)
}
