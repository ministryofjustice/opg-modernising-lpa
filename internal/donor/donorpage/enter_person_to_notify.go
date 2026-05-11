package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type enterPersonToNotifyData struct {
	App  appcontext.Data
	Form *enterPersonToNotifyForm
}

func EnterPersonToNotify(tmpl template.Template, service PeopleToNotifyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) >= 5 {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		addAnother := r.FormValue("addAnother") == "1"
		personToNotify, personFound := provided.PeopleToNotify.Get(actoruid.FromRequest(r))

		if r.Method == http.MethodGet && len(provided.PeopleToNotify) > 0 && personFound == false && addAnother == false {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		data := &enterPersonToNotifyData{
			App:  appData,
			Form: newEnterPersonToNotifyForm(appData.Localizer),
		}

		data.Form.FirstNames.SetInput(personToNotify.FirstNames)
		data.Form.LastName.SetInput(personToNotify.LastName)

		if r.Method == http.MethodPost {
			ok := data.Form.Parse(r)

			nameMatches := personToNotifyMatches(provided, personToNotify.UID, data.Form.FirstNames.Value, data.Form.LastName.Value)
			redirectToWarning := false

			if !nameMatches.IsNone() && personToNotify.NameHasChanged(data.Form.FirstNames.Value, data.Form.LastName.Value) {
				redirectToWarning = true
			}

			if ok {
				personToNotify.FirstNames = data.Form.FirstNames.Value
				personToNotify.LastName = data.Form.LastName.Value

				uid, err := service.Put(r.Context(), provided, personToNotify)
				if err != nil {
					return err
				}

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"id":          {uid.String()},
						"warningFrom": {appData.Page},
						"next": {donor.PathEnterPersonToNotifyAddress.FormatQuery(
							provided.LpaID,
							url.Values{"id": {uid.String()}}),
						},
						"actor": {actor.TypePersonToNotify.String()},
					})
				}

				return donor.PathEnterPersonToNotifyAddress.RedirectQuery(w, r, appData, provided, url.Values{"id": {uid.String()}})
			}
		}

		return tmpl(w, data)
	}
}

type enterPersonToNotifyForm struct {
	newforms.Form
	FirstNames *newforms.String
	LastName   *newforms.String
}

func newEnterPersonToNotifyForm(l Localizer) *enterPersonToNotifyForm {
	return &enterPersonToNotifyForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
	}
}

func (f *enterPersonToNotifyForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.FirstNames, f.LastName)
}
