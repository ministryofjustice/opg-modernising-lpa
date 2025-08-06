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
	PathAreYouSureYouNoLongerNeedVoucher                     = Path("/are-you-sure-you-no-longer-need-voucher")
	PathBecauseYouHaveChosenJointly                          = Path("/because-you-have-chosen-jointly")
	PathBecauseYouHaveChosenJointlyForSomeSeverallyForOthers = Path("/because-you-have-chosen-jointly-for-some-severally-for-others")
	PathCanYouSignYourLpa                                    = Path("/can-you-sign-your-lpa")
	PathCertificateProviderAddress                           = Path("/certificate-provider-address")
	PathCertificateProviderDetails                           = Path("/certificate-provider-details")
	PathCertificateProviderOptOut                            = Path("/certificate-provider-opt-out")
	PathCertificateProviderSummary                           = Path("/certificate-provider-summary")
	PathChangeCertificateProviderMobileNumber                = Path("/change-certificate-provider-mobile-number")
	PathChangeDonorEmail                                     = Path("/change-donor-email")
	PathChangeDonorMobileNumber                              = Path("/change-donor-mobile-number")
	PathChangeIndependentWitnessMobileNumber                 = Path("/change-independent-witness-mobile-number")
	PathCheckYouCanSign                                      = Path("/check-you-can-sign")
	PathCheckYourDetails                                     = Path("/check-your-details")
	PathCheckYourLpa                                         = Path("/check-your-lpa")
	PathChooseAttorneys                                      = Path("/choose-attorneys")
	PathChooseAttorneysAddress                               = Path("/choose-attorneys-address")
	PathChooseAttorneysGuidance                              = Path("/choose-attorneys-guidance")
	PathChooseAttorneysSummary                               = Path("/choose-attorneys-summary")
	PathChooseCertificateProvider                            = Path("/choose-certificate-provider")
	PathChooseCorrespondent                                  = Path("/choose-correspondent")
	PathChooseNewCertificateProvider                         = Path("/choose-new-certificate-provider")
	PathChoosePeopleToNotify                                 = Path("/choose-people-to-notify")
	PathChoosePeopleToNotifySummary                          = Path("/choose-people-to-notify-summary")
	PathChooseReplacementAttorneys                           = Path("/choose-replacement-attorneys")
	PathChooseReplacementAttorneysAddress                    = Path("/choose-replacement-attorneys-address")
	PathChooseReplacementAttorneysSummary                    = Path("/choose-replacement-attorneys-summary")
	PathChooseReplacementTrustCorporation                    = Path("/choose-replacement-trust-corporation")
	PathChooseSomeoneToVouchForYou                           = Path("/choose-someone-to-vouch-for-you")
	PathChooseTrustCorporation                               = Path("/choose-trust-corporation")
	PathChooseYourCertificateProvider                        = Path("/choose-your-certificate-provider")
	PathCompletingYourIdentityConfirmation                   = Path("/completing-your-identity-confirmation")
	PathConfirmPersonAllowedToVouch                          = Path("/confirm-person-allowed-to-vouch")
	PathConfirmYourCertificateProviderIsNotRelated           = Path("/confirm-your-certificate-provider-is-not-related")
	PathConfirmYourIdentity                                  = Path("/confirm-your-identity")
	PathCorrespondentSummary                                 = Path("/correspondent-summary")
	PathCostOfRepeatApplication                              = Path("/cost-of-repeat-application")
	PathDeleteThisLpa                                        = Path("/delete-this-lpa")
	PathDoYouLiveInTheUK                                     = Path("/do-you-live-in-the-uk")
	PathDoYouWantReplacementAttorneys                        = Path("/do-you-want-replacement-attorneys")
	PathDoYouWantToNotifyPeople                              = Path("/do-you-want-to-notify-people")
	PathEnterAttorney                                        = Path("/enter-attorney")
	PathEnterCorrespondentAddress                            = Path("/enter-correspondent-address")
	PathEnterCorrespondentDetails                            = Path("/enter-correspondent-details")
	PathEnterPersonToNotify                                  = Path("/enter-person-to-notify")
	PathEnterPersonToNotifyAddress                           = Path("/enter-person-to-notify-address")
	PathEnterReplacementAttorney                             = Path("/enter-replacement-attorney")
	PathEnterReplacementTrustCorporation                     = Path("/enter-replacement-trust-corporation")
	PathEnterReplacementTrustCorporationAddress              = Path("/enter-replacement-trust-corporation-address")
	PathEnterTrustCorporation                                = Path("/enter-trust-corporation")
	PathEnterTrustCorporationAddress                         = Path("/enter-trust-corporation-address")
	PathEnterVoucher                                         = Path("/enter-voucher")
	PathEvidenceRequired                                     = Path("/evidence-required")
	PathEvidenceSuccessfullyUploaded                         = Path("/evidence-successfully-uploaded")
	PathGettingHelpSigning                                   = Path("/getting-help-signing")
	PathHowDoYouKnowYourCertificateProvider                  = Path("/how-do-you-know-your-certificate-provider")
	PathHowLongHaveYouKnownCertificateProvider               = Path("/how-long-have-you-known-certificate-provider")
	PathHowShouldAttorneysMakeDecisions                      = Path("/how-should-attorneys-make-decisions")
	PathHowShouldReplacementAttorneysMakeDecisions           = Path("/how-should-replacement-attorneys-make-decisions")
	PathHowShouldReplacementAttorneysStepIn                  = Path("/how-should-replacement-attorneys-step-in")
	PathHowToSendEvidence                                    = Path("/how-to-send-evidence")
	PathHowToSignYourLpa                                     = Path("/how-to-sign-your-lpa")
	PathHowWillYouConfirmYourIdentity                        = Path("/how-will-you-confirm-your-identity")
	PathHowWouldCertificateProviderPreferToCarryOutTheirRole = Path("/how-would-certificate-provider-prefer-to-carry-out-their-role")
	PathHowWouldYouLikeToSendEvidence                        = Path("/how-would-you-like-to-send-evidence")
	PathIdentityDetails                                      = Path("/identity-details")
	PathIdentityDetailsUpdated                               = Path("/identity-details-updated")
	PathIdentityWithOneLogin                                 = Path("/id/one-login")
	PathIdentityWithOneLoginCallback                         = Path("/id/one-login/callback")
	PathLifeSustainingTreatment                              = Path("/life-sustaining-treatment")
	PathLpaDetailsSaved                                      = Path("/lpa-details-saved")
	PathLpaType                                              = Path("/lpa-type")
	PathLpaYourLegalRightsAndResponsibilities                = Path("/your-legal-rights-and-responsibilities")
	PathMakeANewLPA                                          = Path("/make-a-new-lpa")
	PathNeedHelpSigningConfirmation                          = Path("/need-help-signing-confirmation")
	PathPayFee                                               = Path("/pay-fee")
	PathPaymentConfirmation                                  = Path("/payment-confirmation")
	PathPaymentSuccessful                                    = Path("/payment-successful")
	PathPendingPayment                                       = Path("/pending-payment")
	PathPreviousApplicationNumber                            = Path("/previous-application-number")
	PathPreviousFee                                          = Path("/how-much-did-you-previously-pay-for-your-lpa")
	PathProgress                                             = Path("/progress")
	PathReadYourLpa                                          = Path("/read-your-lpa")
	PathReceivingUpdatesAboutYourLpa                         = Path("/receiving-updates-about-your-lpa")
	PathRegisterWithCourtOfProtection                        = Path("/register-with-court-of-protection")
	PathRemoveAttorney                                       = Path("/remove-attorney")
	PathRemoveCertificateProvider                            = Path("/remove-certificate-provider")
	PathRemoveCorrespondent                                  = Path("/remove-correspondent")
	PathRemovePersonToNotify                                 = Path("/remove-person-to-notify")
	PathRemoveReplacementAttorney                            = Path("/remove-replacement-attorney")
	PathRemoveReplacementTrustCorporation                    = Path("/remove-replacement-trust-corporation")
	PathRemoveTrustCorporation                               = Path("/remove-trust-corporation")
	PathResendCertificateProviderCode                        = Path("/resend-certificate-provider-code")
	PathResendIndependentWitnessCode                         = Path("/resend-independent-witness-code")
	PathResendVoucherAccessCode                              = Path("/resend-voucher-access-code")
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
	PathWarningInterruption                                  = Path("/warning")
	PathWeHaveContactedVoucher                               = Path("/we-have-contacted-voucher")
	PathWeHaveInformedVoucherNoLongerNeeded                  = Path("/we-have-informed-voucher-no-longer-needed")
	PathWeHaveUpdatedYourDetails                             = Path("/we-have-updated-your-details")
	PathWhatACertificateProviderDoes                         = Path("/what-a-certificate-provider-does")
	PathWhatCountryDoYouLiveIn                               = Path("/what-country-do-you-live-in")
	PathWhatHappensNextRegisteringWithCourtOfProtection      = Path("/what-happens-next-registering-with-court-of-protection")
	PathWhatHappensNextRepeatApplicationNoFee                = Path("/what-happens-next-repeat-application-no-fee")
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
	PathYouHaveToldUsYouAreUnder18                           = Path("/you-have-told-us-you-are-under-18")
	PathYouMustBeOver18ToComplete                            = Path("/you-must-be-over-18-to-complete")
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
	PathYourNonUKAddress                                     = Path("/your-non-uk-address")
	PathYourPreferredLanguage                                = Path("/your-preferred-language")
)

type Path string

func (p Path) String() string {
	return "/lpa/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/lpa/" + id + string(p)
}

func (p Path) FormatQuery(id string, query url.Values) string {
	if len(query) > 0 {
		return p.Format(id) + "?" + query.Encode()
	}

	return p.Format(id)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, donor *donordata.Provided) error {
	rurl := p.Format(donor.LpaID)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, donor.LpaID) {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData appcontext.Data, donor *donordata.Provided, query url.Values) error {
	rurl := p.FormatQuery(donor.LpaID, query)
	if fromURL := r.FormValue("from"); fromURL != "" && canFrom(fromURL, donor.LpaID) && query.Get("warningFrom") == "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p Path) CanGoTo(donor *donordata.Provided) bool {
	if !donor.SignedAt.IsZero() {
		switch p {
		case PathProgress, PathViewLPA, PathDeleteThisLpa, PathWithdrawThisLpa, PathYouHaveSubmittedYourLpa:
			return true

		case PathTaskList:
			return !donor.CompletedAllTasks()

		case PathAboutPayment, PathAreYouApplyingForFeeDiscountOrExemption, PathWhichFeeTypeAreYouApplyingFor,
			PathPreviousApplicationNumber, PathPreviousFee, PathCostOfRepeatApplication, PathEvidenceRequired,
			PathHowWouldYouLikeToSendEvidence, PathUploadEvidence, PathSendUsYourEvidenceByPost, PathPayFee,
			PathPaymentConfirmation, PathPaymentSuccessful, PathEvidenceSuccessfullyUploaded,
			PathWhatHappensNextRepeatApplicationNoFee, PathPendingPayment, PathUploadEvidenceSSE:
			return !donor.Tasks.PayForLpa.IsCompleted()

		case PathConfirmYourIdentity, PathHowWillYouConfirmYourIdentity, PathCompletingYourIdentityConfirmation,
			PathIdentityWithOneLogin, PathIdentityWithOneLoginCallback, PathIdentityDetails, PathRegisterWithCourtOfProtection,
			PathUnableToConfirmIdentity, PathChooseSomeoneToVouchForYou, PathEnterVoucher, PathConfirmPersonAllowedToVouch,
			PathCheckYourDetails, PathWeHaveContactedVoucher, PathWhatYouCanDoNow, PathWhatYouCanDoNowExpired,
			PathWhatHappensNextRegisteringWithCourtOfProtection, PathAreYouSureYouNoLongerNeedVoucher,
			PathWeHaveInformedVoucherNoLongerNeeded:
			return !donor.Tasks.ConfirmYourIdentity.IsCompleted()

		case PathHowToSignYourLpa, PathReadYourLpa, PathYourLpaLanguage, PathLpaYourLegalRightsAndResponsibilities,
			PathSignYourLpa, PathSignTheLpaOnBehalf, PathWitnessingYourSignature, PathWitnessingAsIndependentWitness,
			PathResendIndependentWitnessCode, PathChangeIndependentWitnessMobileNumber, PathWitnessingAsCertificateProvider,
			PathResendCertificateProviderCode, PathChangeCertificateProviderMobileNumber,
			PathCertificateProviderDetails, PathCertificateProviderAddress, PathYourIndependentWitness, PathYourIndependentWitnessAddress:
			return !donor.Tasks.SignTheLpa.IsCompleted()
		}

		return false
	}

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

	case PathHowToSignYourLpa,
		PathReadYourLpa,
		PathSignYourLpa,
		PathWitnessingYourSignature,
		PathWitnessingAsCertificateProvider,
		PathWitnessingAsIndependentWitness,
		PathYouHaveSubmittedYourLpa:
		return section1Completed &&
			(donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending()) &&
			(donor.DonorIdentityConfirmed() || donor.Tasks.ConfirmYourIdentity.IsPending() || donor.RegisteringWithCourtOfProtection || donor.Voucher.FirstNames != "")

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

	case PathConfirmYourIdentity,
		PathHowWillYouConfirmYourIdentity,
		PathIdentityWithOneLogin,
		PathLpaYourLegalRightsAndResponsibilities,
		PathSignTheLpaOnBehalf:
		return section1Completed && (donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending())

	case PathIdentityDetails:
		return (section1Completed || donor.Tasks.CheckYourLpa.IsInProgress()) &&
			(donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending())

	case PathYourName, PathYourDateOfBirth:
		return donor.CanChangePersonalDetails()

	case PathViewLPA:
		return false

	case "":
		return false

	default:
		return true
	}
}

func canFrom(fromURL string, lpaID string) bool {
	return strings.HasPrefix(fromURL, Path("").Format(lpaID))
}
