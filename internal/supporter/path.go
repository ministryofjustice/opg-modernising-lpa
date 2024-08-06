package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

const (
	PathConfirmDonorCanInteractOnline = Path("/confirm-donor-can-interact-online")
	PathContactOPGForPaperForms       = Path("/contact-opg-for-paper-forms")
	PathDashboard                     = Path("/dashboard")
	PathDeleteOrganisation            = Path("/manage-organisation/organisation-details/delete-organisation")
	PathEditMember                    = Path("/manage-organisation/manage-team-members/edit-team-member")
	PathEditOrganisationName          = Path("/manage-organisation/organisation-details/edit-organisation-name")
	PathInviteMember                  = Path("/invite-member")
	PathInviteMemberConfirmation      = Path("/invite-member-confirmation")
	PathManageTeamMembers             = Path("/manage-organisation/manage-team-members")
	PathOrganisationCreated           = Path("/organisation-or-company-created")
	PathOrganisationDetails           = Path("/manage-organisation/organisation-details")

	PathDonorAccess = LpaPath("/donor-access")
	PathViewLPA     = LpaPath("/view-lpa")
)

type Path string

func (p Path) String() string {
	return "/supporter" + string(p)
}

func (p Path) Format() string {
	return "/supporter" + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format()), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format())+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p Path) IsManageOrganisation() bool {
	return p == PathOrganisationDetails ||
		p == PathEditOrganisationName ||
		p == PathManageTeamMembers ||
		p == PathEditMember
}

type LpaPath string

func (p LpaPath) String() string {
	return "/supporter" + string(p) + "/{id}"
}

func (p LpaPath) Format(lpaID string) string {
	return "/supporter" + string(p) + "/" + lpaID
}

func (p LpaPath) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID)), http.StatusFound)
	return nil
}

func (p LpaPath) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID))+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p LpaPath) IsManageOrganisation() bool {
	return false
}
