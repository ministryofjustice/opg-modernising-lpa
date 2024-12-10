package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App      appcontext.Data
	Errors   validation.List
	Provided *attorneydata.Provided
	Lpa      *lpadata.Lpa
	Items    []taskListItem
}

type taskListItem struct {
	Name  string
	Path  attorney.Path
	Query string
	State task.State
	Count int
}

func TaskList(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided, lpa *lpadata.Lpa) error {
		tasks := provided.Tasks

		signItems := []taskListItem{{
			Name:  "signTheLpa",
			Path:  attorney.PathRightsAndResponsibilities,
			State: tasks.SignTheLpa,
		}}

		if provided.WouldLikeSecondSignatory.IsYes() {
			signItems = []taskListItem{{
				Name:  "signTheLpaSignatory1",
				Path:  attorney.PathRightsAndResponsibilities,
				State: tasks.SignTheLpa,
			}, {
				Name:  "signTheLpaSignatory2",
				Path:  attorney.PathSign,
				Query: "?second",
				State: tasks.SignTheLpaSecond,
			}}
		}

		confirmYourDetailsPath := attorney.PathPhoneNumber
		if _, mobile, _ := lpa.Attorney(provided.UID); mobile != "" {
			confirmYourDetailsPath = attorney.PathYourPreferredLanguage
		}
		if tasks.ConfirmYourDetails.IsCompleted() {
			confirmYourDetailsPath = attorney.PathConfirmYourDetails
		}

		data := &taskListData{
			App:      appData,
			Provided: provided,
			Lpa:      lpa,
			Items: append([]taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  confirmYourDetailsPath,
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  attorney.PathReadTheLpa,
					State: tasks.ReadTheLpa,
				},
			}, signItems...),
		}

		return tmpl(w, data)
	}
}
