package page

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

const (
	PathAttorneyConfirmDontWantToBeAttorneyLoggedOut                       = Path("/confirm-you-do-not-want-to-be-an-attorney")
	PathAttorneyEnterReferenceNumber                                       = Path("/attorney-enter-reference-number")
	PathAttorneyEnterReferenceNumberOptOut                                 = Path("/attorney-enter-reference-number-opt-out")
	PathAttorneyLogin                                                      = Path("/attorney-login")
	PathAttorneyLoginCallback                                              = Path("/attorney-login-callback")
	PathAttorneyStart                                                      = Path("/attorney-start")
	PathAttorneyYouHaveDecidedNotToBeAttorney                              = Path("/you-have-decided-not-to-be-an-attorney")
	PathCertificateProviderConfirmDontWantToBeCertificateProviderLoggedOut = Path("/confirm-you-do-not-want-to-be-a-certificate-provider")
	PathCertificateProviderEnterReferenceNumber                            = Path("/certificate-provider-enter-reference-number")
	PathCertificateProviderEnterReferenceNumberOptOut                      = Path("/certificate-provider-enter-reference-number-opt-out")
	PathCertificateProviderLogin                                           = Path("/certificate-provider-login")
	PathCertificateProviderLoginCallback                                   = Path("/certificate-provider-login-callback")
	PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider        = Path("/you-have-decided-not-to-be-a-certificate-provider")
	PathHealthCheckDependency                                              = Path("/health-check/dependency")
	PathHealthCheckService                                                 = Path("/health-check/service")
	PathSupporterEnterOrganisationName                                     = Path("/enter-the-name-of-your-organisation-or-company")
	PathSupporterEnterReferenceNumber                                      = Path("/supporter-reference-number")
	PathSupporterEnterYourName                                             = Path("/enter-your-name")
	PathSupporterInviteExpired                                             = Path("/invite-expired")
	PathSupporterLogin                                                     = Path("/supporter-login")
	PathSupporterLoginCallback                                             = Path("/supporter-login-callback")
	PathSupporterOrganisationDeleted                                       = Path("/organisation-deleted")
	PathSupporterSigningInAdvice                                           = Path("/signing-in-with-govuk-one-login")
	PathSupporterStart                                                     = Path("/supporter-start")
	PathVoucherStart                                                       = Path("/voucher-start")
	PathVoucherLogin                                                       = Path("/voucher-login")
	PathVoucherLoginCallback                                               = Path("/voucher-login-callback")
	PathVoucherEnterReferenceNumber                                        = Path("/voucher-enter-reference-number")

	PathAttorneyFixtures            = Path("/fixtures/attorney")
	PathAuthRedirect                = Path("/auth/redirect")
	PathCertificateProviderFixtures = Path("/fixtures/certificate-provider")
	PathCertificateProviderStart    = Path("/certificate-provider-start")
	PathCookiesConsent              = Path("/cookies-consent")
	PathDashboard                   = Path("/dashboard")
	PathDashboardFixtures           = Path("/fixtures/dashboard")
	PathEnterAccessCode             = Path("/enter-access-code")
	PathFixtures                    = Path("/fixtures")
	PathLogin                       = Path("/login")
	PathLoginCallback               = Path("/login-callback")
	PathLpaDeleted                  = Path("/lpa-deleted")
	PathLpaWithdrawn                = Path("/lpa-withdrawn")
	PathRoot                        = Path("/")
	PathSignOut                     = Path("/sign-out")
	PathStart                       = Path("/start")
	PathSupporterFixtures           = Path("/fixtures/supporter")
	PathVoucherFixtures             = Path("/fixtures/voucher")
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (p Path) Format() string {
	return string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format()), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format())+"?"+query.Encode(), http.StatusFound)
	return nil
}
