package attorney

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
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
	rurl := p.Format(lpaID)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, lpaID) {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string, query url.Values) error {
	rurl := p.Format(lpaID)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, lpaID) {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl)+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p Path) CanGoTo(attorney *attorneydata.Provided, lpa *lpadata.Lpa) bool {
	switch p {
	case PathRightsAndResponsibilities,
		PathWhatHappensWhenYouSign,
		PathSign,
		PathWhatHappensNext:
		return lpa.Paid && lpa.SignedForDonor() &&
			lpa.CertificateProvider.SignedAt != nil && !lpa.CertificateProvider.SignedAt.IsZero() &&
			attorney.Tasks.ConfirmYourDetails.IsCompleted() && attorney.Tasks.ReadTheLpa.IsCompleted()

	case PathWouldLikeSecondSignatory:
		return lpa.Paid && lpa.SignedForDonor() &&
			lpa.CertificateProvider.SignedAt != nil && !lpa.CertificateProvider.SignedAt.IsZero() &&
			attorney.Tasks.ConfirmYourDetails.IsCompleted() && attorney.Tasks.ReadTheLpa.IsCompleted() &&
			attorney.IsTrustCorporation

	default:
		return true
	}
}

func canFrom(fromURL string, lpaID string) bool {
	return strings.HasPrefix(fromURL, Path("").Format(lpaID))
}
