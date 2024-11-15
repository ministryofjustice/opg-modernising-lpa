package voucher

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

const (
	PathConfirmAllowedToVouch         = Path("/confirm-allowed-to-vouch")
	PathConfirmYourIdentity           = Path("/confirm-your-identity")
	PathConfirmYourName               = Path("/confirm-your-name")
	PathDonorDetailsDoNotMatch        = Path("/donor-details-do-not-match")
	PathHowWillYouConfirmYourIdentity = Path("/how-will-you-confirm-your-identity")
	PathIdentityWithOneLogin          = Path("/identity-with-one-login")
	PathIdentityWithOneLoginCallback  = Path("/identity-with-one-login-callback")
	PathOneLoginIdentityDetails       = Path("/one-login-identity-details")
	PathSignTheDeclaration            = Path("/sign-the-declaration")
	PathTaskList                      = Path("/task-list")
	PathThankYou                      = Path("/thank-you")
	PathVerifyDonorDetails            = Path("/verify-donor-details")
	PathYouCannotVouchForDonor        = Path("/you-cannot-vouch-for-donor")
	PathYourName                      = Path("/your-name")
)

type Path string

func (p Path) String() string {
	return "/voucher/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/voucher/" + id + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string) error {
	rurl := p.Format(lpaID)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, lpaID) {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) CanGoTo(provided *voucherdata.Provided) bool {
	switch p {
	case PathYourName:
		return provided.Tasks.ConfirmYourIdentity.IsNotStarted()

	case PathVerifyDonorDetails:
		return provided.Tasks.ConfirmYourName.IsCompleted() &&
			!provided.Tasks.VerifyDonorDetails.IsCompleted()

	case PathConfirmYourIdentity,
		PathHowWillYouConfirmYourIdentity,
		PathIdentityWithOneLogin,
		PathOneLoginIdentityDetails:
		return provided.Tasks.ConfirmYourName.IsCompleted() &&
			provided.Tasks.VerifyDonorDetails.IsCompleted()

	case PathSignTheDeclaration:
		return provided.Tasks.ConfirmYourName.IsCompleted() &&
			provided.Tasks.VerifyDonorDetails.IsCompleted() &&
			provided.Tasks.ConfirmYourIdentity.IsCompleted()

	default:
		return true
	}
}

func canFrom(fromURL string, lpaID string) bool {
	return strings.HasPrefix(fromURL, Path("").Format(lpaID))
}
