package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type manageTeamMembersData struct {
	App            page.AppData
	Errors         validation.List
	Organisation   *actor.Organisation
	InvitedMembers []*actor.MemberInvite
	Members        []*actor.Member
}

func ManageTeamMembers(tmpl template.Template, memberStore MemberStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		invitedMembers, err := memberStore.InvitedMembers(r.Context())
		if err != nil {
			return err
		}

		members, err := memberStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &manageTeamMembersData{
			App:            appData,
			Organisation:   organisation,
			InvitedMembers: invitedMembers,
			Members:        members,
		})
	}
}
