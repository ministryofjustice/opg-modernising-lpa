package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type WarningInterruptionData struct {
	App           appcontext.Data
	Errors        validation.List
	Donor         *donordata.Provided
	Attorney      donordata.Attorney
	Notifications []page.Notification
	PageTitle     string
	From          string
}

func WarningInterruption(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := WarningInterruptionData{
			App:   appData,
			Donor: provided,
			From:  r.FormValue("warningFrom"),
		}

		switch data.From {
		case donor.PathChooseAttorneys.Format(appData.LpaID):
			uid, err := actoruid.Parse(r.FormValue("id"))
			if err != nil {
				return donor.PathTaskList.RedirectQuery(w, r, appData, provided, nil)
			}

			attorney, found := provided.Attorneys.Get(uid)

			if found {
				data.Attorney = attorney

				nameWarning := actor.NewSameNameWarning(
					actor.TypeAttorney,
					attorneyMatches(provided, uid, attorney.FirstNames, attorney.LastName),
					attorney.FirstNames,
					attorney.LastName,
				)
				dobWarning := DateOfBirthWarning(attorney.DateOfBirth, false)

				if dobWarning != "" {
					data.Notifications = append(data.Notifications, page.Notification{
						Heading:  "pleaseReviewTheInformationYouHaveEntered",
						BodyHTML: dobWarning,
					})
				}

				if nameWarning != nil {
					data.Notifications = append(data.Notifications, page.Notification{
						Heading:  "pleaseReviewTheInformationYouHaveEntered",
						BodyHTML: nameWarning.Format(appData.Localizer),
					})
				}

				data.PageTitle = "checkYourAttorneysDetails"
			}
		}

		if len(data.Notifications) == 0 {
			return donor.PathTaskList.RedirectQuery(w, r, appData, provided, nil)
		}

		return tmpl(w, data)
	}
}

func attorneyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsAttorney() && person.UID == uid) &&
			strings.EqualFold(person.FirstNames, firstNames) &&
			strings.EqualFold(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func DateOfBirthWarning(dateOfBirth date.Date, replacement bool) string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !dateOfBirth.IsZero() {
		if dateOfBirth.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if dateOfBirth.Before(today) && dateOfBirth.After(eighteenYearsEarlier) {
			//TODO drop as part of MLPAB-2990
			if replacement {
				return "attorneyDateOfBirthIsUnder18"
			}
			return "dateOfBirthIsUnder18Attorney"
		}
	}

	return ""
}
