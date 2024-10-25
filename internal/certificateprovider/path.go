package certificateprovider

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
)

const (
	PathCertificateProvided                    = Path("/certificate-provided")
	PathConfirmDontWantToBeCertificateProvider = Path("/confirm-you-do-not-want-to-be-a-certificate-provider")
	PathConfirmYourDetails                     = Path("/confirm-your-details")
	PathEnterDateOfBirth                       = Path("/enter-date-of-birth")
	PathIdentityWithOneLogin                   = Path("/identity-with-one-login")
	PathIdentityWithOneLoginCallback           = Path("/identity-with-one-login-callback")
	PathOneLoginIdentityDetails                = Path("/one-login-identity-details")
	PathConfirmYourIdentity                    = Path("/confirm-your-identity")
	PathProvideCertificate                     = Path("/provide-certificate")
	PathReadTheLpa                             = Path("/read-the-lpa")
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
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) canVisit(certificateProvider *certificateproviderdata.Provided) bool {
	switch p {
	case PathConfirmYourIdentity,
		PathIdentityWithOneLogin,
		PathIdentityWithOneLoginCallback:
		return certificateProvider.Tasks.ConfirmYourDetails.IsCompleted()

	case PathWhatHappensNext,
		PathProvideCertificate,
		PathConfirmDontWantToBeCertificateProvider,
		PathCertificateProvided:
		return certificateProvider.Tasks.ConfirmYourDetails.IsCompleted() && certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted()

	default:
		return true
	}
}

func CanGoTo(certificateProvider *certificateproviderdata.Provided, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/certificate-provider/") {
		_, certificateProviderPath, _ := strings.Cut(strings.TrimPrefix(path, "/certificate-provider/"), "/")
		return Path("/" + certificateProviderPath).canVisit(certificateProvider)
	}

	return true
}
