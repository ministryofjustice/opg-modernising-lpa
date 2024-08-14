package voucher

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

const (
	PathTaskList            = Path("/task-list")
	PathConfirmYourName     = Path("/confirm-your-name")
	PathYourName            = Path("/your-name")
	PathVerifyDonorDetails  = Path("/verify-donor-details")
	PathConfirmYourIdentity = Path("/confirm-your-identity")
	PathSignTheDeclaration  = Path("/sign-the-declaration")
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

func (p Path) canVisit(provided *voucherdata.Provided) bool {
	switch p {
	case PathVerifyDonorDetails:
		return provided.Tasks.ConfirmYourName.Completed()

	case PathConfirmYourIdentity:
		return provided.Tasks.ConfirmYourName.Completed() &&
			provided.Tasks.VerifyDonorDetails.Completed()

	case PathSignTheDeclaration:
		return provided.Tasks.ConfirmYourName.Completed() &&
			provided.Tasks.VerifyDonorDetails.Completed() &&
			provided.Tasks.ConfirmYourIdentity.Completed()

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
		return Path("/" + voucherPath).canVisit(provided)
	}

	return true
}
