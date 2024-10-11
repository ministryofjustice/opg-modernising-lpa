package donor

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

const (
	PathAboutPayment                                         = Path("/about-payment")
	PathAddCorrespondent                                     = Path("/add-correspondent")
	PathAreYouApplyingForFeeDiscountOrExemption              = Path("/are-you-applying-for-fee-discount-or-exemption")
	PathBecauseYouHaveChosenJointly                          = Path("/because-you-have-chosen-jointly")
	PathBecauseYouHaveChosenJointlyForSomeSeverallyForOthers = Path("/because-you-have-chosen-jointly-for-some-severally-for-others")
	PathCanYouSignYourLpa                                    = Path("/can-you-sign-your-lpa")
	PathCertificateProviderAddress                           = Path("/certificate-provider-address")
	PathCertificateProviderDetails                           = Path("/certificate-provider-details")
	PathCertificateProviderOptOut                            = Path("/certificate-provider-opt-out")
	PathChangeCertificateProviderMobileNumber                = Path("/change-certificate-provider-mobile-number")
	PathChangeIndependentWitnessMobileNumber                 = Path("/change-independent-witness-mobile-number")
	PathCheckYouCanSign                                      = Path("/check-you-can-sign")
	PathCheckYourDetails                                     = Path("/check-your-details")
	PathCheckYourLpa                                         = Path("/check-your-lpa")
	PathChooseAttorneys                                      = Path("/choose-attorneys")
	PathChooseAttorneysAddress                               = Path("/choose-attorneys-address")
	PathChooseAttorneysGuidance                              = Path("/choose-attorneys-guidance")
	PathChooseAttorneysSummary                               = Path("/choose-attorneys-summary")
	PathChooseNewCertificateProvider                         = Path("/choose-new-certificate-provider")
	PathChoosePeopleToNotify                                 = Path("/choose-people-to-notify")
	PathChoosePeopleToNotifyAddress                          = Path("/choose-people-to-notify-address")
	PathChoosePeopleToNotifySummary                          = Path("/choose-people-to-notify-summary")
	PathChooseReplacementAttorneys                           = Path("/choose-replacement-attorneys")
	PathChooseReplacementAttorneysAddress                    = Path("/choose-replacement-attorneys-address")
	PathChooseReplacementAttorneysSummary                    = Path("/choose-replacement-attorneys-summary")
	PathChooseYourCertificateProvider                        = Path("/choose-your-certificate-provider")
	PathConfirmPersonAllowedToVouch                          = Path("/confirm-person-allowed-to-vouch")
	PathConfirmYourCertificateProviderIsNotRelated           = Path("/confirm-your-certificate-provider-is-not-related")
	PathDeleteThisLpa                                        = Path("/delete-this-lpa")
	PathDoYouWantReplacementAttorneys                        = Path("/do-you-want-replacement-attorneys")
	PathDoYouWantToNotifyPeople                              = Path("/do-you-want-to-notify-people")
	PathEnterCorrespondentAddress                            = Path("/enter-correspondent-address")
	PathEnterCorrespondentDetails                            = Path("/enter-correspondent-details")
	PathEnterReplacementTrustCorporation                     = Path("/enter-replacement-trust-corporation")
	PathEnterReplacementTrustCorporationAddress              = Path("/enter-replacement-trust-corporation-address")
	PathEnterTrustCorporation                                = Path("/enter-trust-corporation")
	PathEnterTrustCorporationAddress                         = Path("/enter-trust-corporation-address")
	PathEnterVoucher                                         = Path("/enter-voucher")
	PathEvidenceRequired                                     = Path("/evidence-required")
	PathEvidenceSuccessfullyUploaded                         = Path("/evidence-successfully-uploaded")
	PathFeeApproved                                          = Path("/fee-approved")
	PathFeeDenied                                            = Path("/fee-denied")
	PathGettingHelpSigning                                   = Path("/getting-help-signing")
	PathHowDoYouKnowYourCertificateProvider                  = Path("/how-do-you-know-your-certificate-provider")
	PathHowLongHaveYouKnownCertificateProvider               = Path("/how-long-have-you-known-certificate-provider")
	PathHowShouldAttorneysMakeDecisions                      = Path("/how-should-attorneys-make-decisions")
	PathHowShouldReplacementAttorneysMakeDecisions           = Path("/how-should-replacement-attorneys-make-decisions")
	PathHowShouldReplacementAttorneysStepIn                  = Path("/how-should-replacement-attorneys-step-in")
	PathHowToConfirmYourIdentityAndSign                      = Path("/how-to-confirm-your-identity-and-sign")
	PathHowToSendEvidence                                    = Path("/how-to-send-evidence")
	PathHowWouldCertificateProviderPreferToCarryOutTheirRole = Path("/how-would-certificate-provider-prefer-to-carry-out-their-role")
	PathHowWouldYouLikeToSendEvidence                        = Path("/how-would-you-like-to-send-evidence")
	PathIdentityWithOneLogin                                 = Path("/id/one-login")
	PathIdentityWithOneLoginCallback                         = Path("/id/one-login/callback")
	PathLifeSustainingTreatment                              = Path("/life-sustaining-treatment")
	PathLpaDetailsSaved                                      = Path("/lpa-details-saved")
	PathLpaType                                              = Path("/lpa-type")
	PathLpaYourLegalRightsAndResponsibilities                = Path("/your-legal-rights-and-responsibilities")
	PathMakeANewLPA                                          = Path("/make-a-new-lpa")
	PathNeedHelpSigningConfirmation                          = Path("/need-help-signing-confirmation")
	PathOneLoginIdentityDetails                              = Path("/one-login-identity-details")
	PathPaymentConfirmation                                  = Path("/payment-confirmation")
	PathPaymentSuccessful                                    = Path("/payment-successful")
	PathPreviousApplicationNumber                            = Path("/previous-application-number")
	PathPreviousFee                                          = Path("/how-much-did-you-previously-pay-for-your-lpa")
	PathProgress                                             = Path("/progress")
	PathProveYourIdentity                                    = Path("/prove-your-identity")
	PathReadYourLpa                                          = Path("/read-your-lpa")
	PathReceivingUpdatesAboutYourLpa                         = Path("/receiving-updates-about-your-lpa")
	PathRegisterWithCourtOfProtection                        = Path("/register-with-court-of-protection")
	PathRemoveAttorney                                       = Path("/remove-attorney")
	PathRemovePersonToNotify                                 = Path("/remove-person-to-notify")
	PathRemoveReplacementAttorney                            = Path("/remove-replacement-attorney")
	PathRemoveReplacementTrustCorporation                    = Path("/remove-replacement-trust-corporation")
	PathRemoveTrustCorporation                               = Path("/remove-trust-corporation")
	PathResendCertificateProviderCode                        = Path("/resend-certificate-provider-code")
	PathResendIndependentWitnessCode                         = Path("/resend-independent-witness-code")
	PathRestrictions                                         = Path("/restrictions")
	PathSendUsYourEvidenceByPost                             = Path("/send-us-your-evidence-by-post")
	PathSignTheLpaOnBehalf                                   = Path("/sign-the-lpa-on-behalf")
	PathSignYourLpa                                          = Path("/sign-your-lpa")
	PathTaskList                                             = Path("/task-list")
	PathUnableToConfirmIdentity                              = Path("/unable-to-confirm-identity")
	PathUploadEvidence                                       = Path("/upload-evidence")
	PathUploadEvidenceSSE                                    = Path("/upload-evidence-sse")
	PathUseExistingAddress                                   = Path("/use-existing-address")
	PathViewLPA                                              = Path("/view-lpa")
	PathWeHaveContactedVoucher                               = Path("/we-have-contacted-voucher")
	PathWeHaveReceivedVoucherDetails                         = Path("/we-have-received-voucher-details")
	PathWeHaveUpdatedYourDetails                             = Path("/we-have-updated-your-details")
	PathWhatACertificateProviderDoes                         = Path("/what-a-certificate-provider-does")
	PathWhatHappensNextPostEvidence                          = Path("/what-happens-next-post-evidence")
	PathWhatHappensNextRegisteringWithCourtOfProtection      = Path("/what-happens-next-registering-with-court-of-protection")
	PathWhatIsVouching                                       = Path("/what-is-vouching")
	PathWhatYouCanDoNow                                      = Path("/what-you-can-do-now")
	PathWhatYouCanDoNowExpired                               = Path("/what-you-can-do-now-expired")
	PathWhenCanTheLpaBeUsed                                  = Path("/when-can-the-lpa-be-used")
	PathWhichFeeTypeAreYouApplyingFor                        = Path("/which-fee-type-are-you-applying-for")
	PathWithdrawThisLpa                                      = Path("/withdraw-this-lpa")
	PathWitnessingAsCertificateProvider                      = Path("/witnessing-as-certificate-provider")
	PathWitnessingAsIndependentWitness                       = Path("/witnessing-as-independent-witness")
	PathWitnessingYourSignature                              = Path("/witnessing-your-signature")
	PathYouCannotSignYourLpaYet                              = Path("/you-cannot-sign-your-lpa-yet")
	PathYouHaveSubmittedYourLpa                              = Path("/you-have-submitted-your-lpa")
	PathYourAddress                                          = Path("/your-address")
	PathYourAuthorisedSignatory                              = Path("/your-authorised-signatory")
	PathYourDateOfBirth                                      = Path("/your-date-of-birth")
	PathYourDetails                                          = Path("/your-details")
	PathYourEmail                                            = Path("/your-email")
	PathYourIndependentWitness                               = Path("/your-independent-witness")
	PathYourIndependentWitnessAddress                        = Path("/your-independent-witness-address")
	PathYourIndependentWitnessMobile                         = Path("/your-independent-witness-mobile")
	PathYourLegalRightsAndResponsibilitiesIfYouMakeLpa       = Path("/your-legal-rights-and-responsibilities-if-you-make-an-lpa")
	PathYourLpaLanguage                                      = Path("/your-lpa-language")
	PathYourMobile                                           = Path("/your-mobile")
	PathYourName                                             = Path("/your-name")
	PathYourPreferredLanguage                                = Path("/your-preferred-language")
)

type Path string

func (p Path) String() string {
	return "/lpa/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/lpa/" + id + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, donor *donordata.Provided) error {
	rurl := p.Format(donor.LpaID)
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, donor *donordata.Provided, query url.Values) error {
	rurl := p.Format(donor.LpaID) + "?" + query.Encode()
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) canVisit(donor *donordata.Provided) bool {
	section1Completed := donor.Tasks.YourDetails.IsCompleted() &&
		donor.Tasks.ChooseAttorneys.IsCompleted() &&
		donor.Tasks.ChooseReplacementAttorneys.IsCompleted() &&
		(donor.Type.IsPersonalWelfare() && donor.Tasks.LifeSustainingTreatment.IsCompleted() || donor.Type.IsPropertyAndAffairs() && donor.Tasks.WhenCanTheLpaBeUsed.IsCompleted()) &&
		donor.Tasks.Restrictions.IsCompleted() &&
		donor.Tasks.CertificateProvider.IsCompleted() &&
		donor.Tasks.PeopleToNotify.IsCompleted() &&
		(donor.Donor.CanSign.IsYes() || donor.Tasks.ChooseYourSignatory.IsCompleted()) &&
		donor.Tasks.CheckYourLpa.IsCompleted()

	switch p {
	case PathWhenCanTheLpaBeUsed,
		PathLifeSustainingTreatment,
		PathRestrictions,
		PathWhatACertificateProviderDoes,
		PathDoYouWantToNotifyPeople,
		PathDoYouWantReplacementAttorneys:
		return donor.Tasks.YourDetails.IsCompleted() && donor.Tasks.ChooseAttorneys.IsCompleted()

	case PathGettingHelpSigning:
		return donor.Tasks.CertificateProvider.IsCompleted()

	case PathReadYourLpa,
		PathSignYourLpa,
		PathWitnessingYourSignature,
		PathWitnessingAsCertificateProvider,
		PathWitnessingAsIndependentWitness,
		PathYouHaveSubmittedYourLpa:
		return section1Completed &&
			(donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending()) &&
			(donor.DonorIdentityConfirmed() || donor.RegisteringWithCourtOfProtection || donor.Voucher.FirstNames != "")

	case PathConfirmYourCertificateProviderIsNotRelated,
		PathCheckYourLpa:
		return donor.Tasks.YourDetails.IsCompleted() &&
			donor.Tasks.ChooseAttorneys.IsCompleted() &&
			donor.Tasks.ChooseReplacementAttorneys.IsCompleted() &&
			(donor.Type.IsPersonalWelfare() && donor.Tasks.LifeSustainingTreatment.IsCompleted() || donor.Tasks.WhenCanTheLpaBeUsed.IsCompleted()) &&
			donor.Tasks.Restrictions.IsCompleted() &&
			donor.Tasks.CertificateProvider.IsCompleted() &&
			donor.Tasks.PeopleToNotify.IsCompleted() &&
			(donor.Donor.CanSign.IsYes() || donor.Tasks.ChooseYourSignatory.IsCompleted()) &&
			donor.Tasks.AddCorrespondent.IsCompleted()

	case PathAboutPayment:
		return section1Completed

	case PathHowToConfirmYourIdentityAndSign,
		PathIdentityWithOneLogin,
		PathOneLoginIdentityDetails,
		PathLpaYourLegalRightsAndResponsibilities,
		PathSignTheLpaOnBehalf:
		return section1Completed && (donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending())

	case PathYourName, PathYourDateOfBirth:
		return donor.IdentityUserData.Status.IsUnknown()

	default:
		return true
	}
}

func CanGoTo(donor *donordata.Provided, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/lpa/") {
		_, lpaPath, _ := strings.Cut(strings.TrimPrefix(path, "/lpa/"), "/")
		return Path("/" + lpaPath).canVisit(donor)
	}

	return true
}
