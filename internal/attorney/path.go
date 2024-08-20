package attorney

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
)

const (
	PathCodeOfConduct               = Path("/code-of-conduct")
	PathConfirmDontWantToBeAttorney = Path("/confirm-you-do-not-want-to-be-an-attorney")
	PathConfirmYourDetails          = Path("/confirm-your-details")
	PathPhoneNumber                 = Path("/phone-number")
	PathProgress                    = Path("/progress")
	PathReadTheLpa                  = Path("/read-the-lpa")
	PathRightsAndResponsibilities   = Path("/legal-rights-and-responsibilities")
	PathSign                        = Path("/sign")
	PathTaskList                    = Path("/task-list")
	PathWhatHappensNext             = Path("/what-happens-next")
	PathWhatHappensWhenYouSign      = Path("/what-happens-when-you-sign-the-lpa")
	PathWouldLikeSecondSignatory    = Path("/would-like-second-signatory")
	PathYourPreferredLanguage       = Path("/your-preferred-language")
)

type Path string

func (p Path) String() string {
	return "/attorney/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/attorney/" + id + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID)), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID))+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p Path) canVisit(attorney *attorneydata.Provided) bool {
	switch p {
	case PathRightsAndResponsibilities,
		PathWhatHappensWhenYouSign,
		PathSign,
		PathWhatHappensNext:
		return attorney.Tasks.ConfirmYourDetails.IsCompleted() && attorney.Tasks.ReadTheLpa.IsCompleted()

	case PathWouldLikeSecondSignatory:
		return attorney.Tasks.ConfirmYourDetails.IsCompleted() && attorney.Tasks.ReadTheLpa.IsCompleted() && attorney.IsTrustCorporation

	default:
		return true
	}
}

func CanGoTo(attorney *attorneydata.Provided, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/attorney/") {
		_, attorneyPath, _ := strings.Cut(strings.TrimPrefix(path, "/attorney/"), "/")
		return Path("/" + attorneyPath).canVisit(attorney)
	}

	return true
}
