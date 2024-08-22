package voucher

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

const (
	PathTaskList                     = Path("/task-list")
	PathConfirmAllowedToVouch        = Path("/confirm-allowed-to-vouch")
	PathConfirmYourName              = Path("/confirm-your-name")
	PathYourName                     = Path("/your-name")
	PathVerifyDonorDetails           = Path("/verify-donor-details")
	PathDonorDetailsDoNotMatch       = Path("/donor-details-do-not-match")
	PathConfirmYourIdentity          = Path("/confirm-your-identity")
	PathSignTheDeclaration           = Path("/sign-the-declaration")
	PathIdentityWithOneLogin         = Path("/identity-with-one-login")
	PathIdentityWithOneLoginCallback = Path("/identity-with-one-login-callback")
	PathOneLoginIdentityDetails      = Path("/one-login-identity-details")
	PathUnableToConfirmIdentity      = Path("/unable-to-confirm-identity")
	PathThankYou                     = Path("/thank-you")
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
	if fromURL := r.FormValue("from"); fromURL != "" {
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

	case PathConfirmYourIdentity:
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

func CanGoTo(provided *voucherdata.Provided, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/voucher/") {
		_, voucherPath, _ := strings.Cut(strings.TrimPrefix(path, "/voucher/"), "/")
		return Path("/" + voucherPath).CanGoTo(provided)
	}

	return true
}
