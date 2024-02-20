package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type organisationDetailsData struct {
	App            page.AppData
	Query          url.Values
	Errors         validation.List
	Organisation   *actor.Organisation
	InvitedMembers []*actor.MemberInvite
}

func OrganisationDetails(tmpl template.Template, memberStore MemberStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		invitedMembers, err := memberStore.InvitedMembers(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &organisationDetailsData{
			App:            appData,
			Query:          r.URL.Query(),
			Organisation:   organisation,
			InvitedMembers: invitedMembers,
		})
	}
}
