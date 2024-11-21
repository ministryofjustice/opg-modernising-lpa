package certificateprovider

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

const (
	PathCertificateProvided                    = Path("/certificate-provided")
	PathCompletingYourIdentityConfirmation     = Path("/completing-your-identity-confirmation")
	PathConfirmDontWantToBeCertificateProvider = Path("/confirm-you-do-not-want-to-be-a-certificate-provider")
	PathConfirmYourDetails                     = Path("/confirm-your-details")
	PathConfirmYourIdentity                    = Path("/confirm-your-identity")
	PathEnterDateOfBirth                       = Path("/enter-date-of-birth")
	PathHowWillYouConfirmYourIdentity          = Path("/how-will-you-confirm-your-identity")
	PathIdentityWithOneLogin                   = Path("/identity-with-one-login")
	PathIdentityWithOneLoginCallback           = Path("/identity-with-one-login-callback")
	PathOneLoginIdentityDetails                = Path("/one-login-identity-details")
	PathProvideCertificate                     = Path("/provide-certificate")
	PathReadTheLpa                             = Path("/read-the-lpa")
	PathReadTheDraftLpa                        = Path("/read-the-draft-lpa")
	PathTaskList                               = Path("/task-list")
	PathUnableToConfirmIdentity                = Path("/unable-to-confirm-identity")
	PathWhatHappensNext                        = Path("/what-happens-next")
	PathWhatIsYourHomeAddress                  = Path("/what-is-your-home-address")
	PathWhoIsEligible                          = Path("/certificate-provider-who-is-eligible")
	PathYourPreferredLanguage                  = Path("/your-preferred-language")
	PathYourRole                               = Path("/your-role")
)

type Path string

func (p Path) String() string {
	return "/certificate-provider/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/certificate-provider/" + id + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string) error {
	rurl := p.Format(lpaID)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, lpaID) {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) CanGoTo(certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) bool {
	switch p {
	case PathConfirmYourIdentity,
		PathHowWillYouConfirmYourIdentity,
		PathIdentityWithOneLogin,
		PathIdentityWithOneLoginCallback,
		PathOneLoginIdentityDetails:
		return lpa.Paid && lpa.SignedForDonor() &&
			certificateProvider.Tasks.ConfirmYourDetails.IsCompleted()

	case PathWhatHappensNext,
		PathReadTheLpa,
		PathProvideCertificate,
		PathConfirmDontWantToBeCertificateProvider,
		PathCertificateProvided:
		return lpa.Paid && lpa.SignedForDonor() &&
			certificateProvider.Tasks.ConfirmYourDetails.IsCompleted() &&
			(certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted() || certificateProvider.Tasks.ConfirmYourIdentity.IsPending())

	default:
		return true
	}
}

func canFrom(fromURL string, lpaID string) bool {
	return strings.HasPrefix(fromURL, Path("").Format(lpaID))
}
