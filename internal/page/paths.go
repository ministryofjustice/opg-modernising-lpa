package page

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
)

type Path string

func (p Path) String() string {
	return string(p)
}

func (p Path) Format() string {
	return string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData AppData) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format()), http.StatusFound)
	return nil
}

func (p Path) RedirectQuery(w http.ResponseWriter, r *http.Request, appData AppData, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format())+"?"+query.Encode(), http.StatusFound)
	return nil
}

type LpaPath string

func (p LpaPath) String() string {
	return "/lpa/{id}" + string(p)
}

func (p LpaPath) Format(id string) string {
	return "/lpa/" + id + string(p)
}

func (p LpaPath) Redirect(w http.ResponseWriter, r *http.Request, appData AppData, donor *actor.DonorProvidedDetails) error {
	rurl := p.Format(donor.LpaID)
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p LpaPath) RedirectQuery(w http.ResponseWriter, r *http.Request, appData AppData, donor *actor.DonorProvidedDetails, query url.Values) error {
	rurl := p.Format(donor.LpaID) + "?" + query.Encode()
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}

func (p LpaPath) canVisit(donor *actor.DonorProvidedDetails) bool {
	section1Completed := donor.Tasks.YourDetails.Completed() &&
		donor.Tasks.ChooseAttorneys.Completed() &&
		donor.Tasks.ChooseReplacementAttorneys.Completed() &&
		(donor.Type.IsPersonalWelfare() && donor.Tasks.LifeSustainingTreatment.Completed() || donor.Type.IsPropertyAndAffairs() && donor.Tasks.WhenCanTheLpaBeUsed.Completed()) &&
		donor.Tasks.Restrictions.Completed() &&
		donor.Tasks.CertificateProvider.Completed() &&
		donor.Tasks.PeopleToNotify.Completed() &&
		(donor.Donor.CanSign.IsYes() || donor.Tasks.ChooseYourSignatory.Completed()) &&
		donor.Tasks.CheckYourLpa.Completed()

	switch p {
	case Paths.WhenCanTheLpaBeUsed,
		Paths.LifeSustainingTreatment,
		Paths.Restrictions,
		Paths.WhatACertificateProviderDoes,
		Paths.DoYouWantToNotifyPeople,
		Paths.DoYouWantReplacementAttorneys:
		return donor.Tasks.YourDetails.Completed() && donor.Tasks.ChooseAttorneys.Completed()

	case Paths.GettingHelpSigning:
		return donor.Tasks.CertificateProvider.Completed()

	case Paths.ReadYourLpa,
		Paths.SignYourLpa,
		Paths.WitnessingYourSignature,
		Paths.WitnessingAsCertificateProvider,
		Paths.WitnessingAsIndependentWitness,
		Paths.YouHaveSubmittedYourLpa:
		return donor.DonorIdentityConfirmed()

	case Paths.ConfirmYourCertificateProviderIsNotRelated,
		Paths.CheckYourLpa:
		return donor.Tasks.YourDetails.Completed() &&
			donor.Tasks.ChooseAttorneys.Completed() &&
			donor.Tasks.ChooseReplacementAttorneys.Completed() &&
			(donor.Type.IsPersonalWelfare() && donor.Tasks.LifeSustainingTreatment.Completed() || donor.Tasks.WhenCanTheLpaBeUsed.Completed()) &&
			donor.Tasks.Restrictions.Completed() &&
			donor.Tasks.CertificateProvider.Completed() &&
			donor.Tasks.PeopleToNotify.Completed() &&
			(donor.Donor.CanSign.IsYes() || donor.Tasks.ChooseYourSignatory.Completed()) &&
			donor.Tasks.AddCorrespondent.Completed()

	case Paths.AboutPayment:
		return section1Completed

	case Paths.HowToConfirmYourIdentityAndSign,
		Paths.IdentityWithOneLogin,
		Paths.SignTheLpaOnBehalf:
		return section1Completed && (donor.Tasks.PayForLpa.IsCompleted() || donor.Tasks.PayForLpa.IsPending())

	default:
		return true
	}
}

type AttorneyPath string

func (p AttorneyPath) String() string {
	return "/attorney/{id}" + string(p)
}

func (p AttorneyPath) Format(id string) string {
	return "/attorney/" + id + string(p)
}

func (p AttorneyPath) Redirect(w http.ResponseWriter, r *http.Request, appData AppData, lpaID string) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID)), http.StatusFound)
	return nil
}

func (p AttorneyPath) RedirectQuery(w http.ResponseWriter, r *http.Request, appData AppData, lpaID string, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID))+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p AttorneyPath) canVisit(attorney *actor.AttorneyProvidedDetails) bool {
	switch p {
	case Paths.Attorney.RightsAndResponsibilities,
		Paths.Attorney.WhatHappensWhenYouSign,
		Paths.Attorney.Sign,
		Paths.Attorney.WhatHappensNext:
		return attorney.Tasks.ConfirmYourDetails.Completed() && attorney.Tasks.ReadTheLpa.Completed()

	case Paths.Attorney.WouldLikeSecondSignatory:
		return attorney.Tasks.ConfirmYourDetails.Completed() && attorney.Tasks.ReadTheLpa.Completed() && attorney.IsTrustCorporation

	default:
		return true
	}
}

type CertificateProviderPath string

func (p CertificateProviderPath) String() string {
	return "/certificate-provider/{id}" + string(p)
}

func (p CertificateProviderPath) Format(id string) string {
	return "/certificate-provider/" + id + string(p)
}

func (p CertificateProviderPath) Redirect(w http.ResponseWriter, r *http.Request, appData AppData, lpaID string) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID)), http.StatusFound)
	return nil
}

func (p CertificateProviderPath) canVisit(certificateProvider *actor.CertificateProviderProvidedDetails) bool {
	switch p {
	case Paths.CertificateProvider.ProveYourIdentity,
		Paths.CertificateProvider.IdentityWithOneLogin,
		Paths.CertificateProvider.IdentityWithOneLoginCallback:
		return certificateProvider.Tasks.ConfirmYourDetails.Completed()

	case Paths.CertificateProvider.WhatHappensNext,
		Paths.CertificateProvider.ProvideCertificate,
		Paths.CertificateProvider.ConfirmDontWantToBeCertificateProvider,
		Paths.CertificateProvider.CertificateProvided:
		return certificateProvider.Tasks.ConfirmYourDetails.Completed() && certificateProvider.Tasks.ConfirmYourIdentity.Completed()

	default:
		return true
	}
}

type SupporterPath string

func (p SupporterPath) String() string {
	return "/supporter" + string(p)
}

func (p SupporterPath) Format() string {
	return "/supporter" + string(p)
}

func (p SupporterPath) Redirect(w http.ResponseWriter, r *http.Request, appData AppData) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format()), http.StatusFound)
	return nil
}

func (p SupporterPath) RedirectQuery(w http.ResponseWriter, r *http.Request, appData AppData, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format())+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p SupporterPath) IsManageOrganisation() bool {
	return p == Paths.Supporter.OrganisationDetails ||
		p == Paths.Supporter.EditOrganisationName ||
		p == Paths.Supporter.ManageTeamMembers ||
		p == Paths.Supporter.EditMember
}

type SupporterLpaPath string

func (p SupporterLpaPath) String() string {
	return "/supporter" + string(p) + "/{id}"
}

func (p SupporterLpaPath) Format(lpaID string) string {
	return "/supporter" + string(p) + "/" + lpaID
}

func (p SupporterLpaPath) Redirect(w http.ResponseWriter, r *http.Request, appData AppData, lpaID string) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID)), http.StatusFound)
	return nil
}

func (p SupporterLpaPath) RedirectQuery(w http.ResponseWriter, r *http.Request, appData AppData, lpaID string, query url.Values) error {
	http.Redirect(w, r, appData.Lang.URL(p.Format(lpaID))+"?"+query.Encode(), http.StatusFound)
	return nil
}

func (p SupporterLpaPath) IsManageOrganisation() bool {
	return false
}

type AttorneyPaths struct {
	EnterReferenceNumber Path
	Login                Path
	LoginCallback        Path
	Start                Path

	CodeOfConduct             AttorneyPath
	ConfirmYourDetails        AttorneyPath
	MobileNumber              AttorneyPath
	Progress                  AttorneyPath
	ReadTheLpa                AttorneyPath
	RightsAndResponsibilities AttorneyPath
	Sign                      AttorneyPath
	TaskList                  AttorneyPath
	WhatHappensNext           AttorneyPath
	WhatHappensWhenYouSign    AttorneyPath
	WouldLikeSecondSignatory  AttorneyPath
	YourPreferredLanguage     AttorneyPath
}

type CertificateProviderPaths struct {
	Login                                           Path
	LoginCallback                                   Path
	EnterReferenceNumber                            Path
	EnterReferenceNumberOptOut                      Path
	ConfirmDontWantToBeCertificateProviderLoggedOut Path
	YouHaveDecidedNotToBeACertificateProvider       Path

	CertificateProvided                    CertificateProviderPath
	ConfirmDontWantToBeCertificateProvider CertificateProviderPath
	ConfirmYourDetails                     CertificateProviderPath
	EnterDateOfBirth                       CertificateProviderPath
	IdentityWithOneLogin                   CertificateProviderPath
	IdentityWithOneLoginCallback           CertificateProviderPath
	ProveYourIdentity                      CertificateProviderPath
	ProvideCertificate                     CertificateProviderPath
	ReadTheLpa                             CertificateProviderPath
	TaskList                               CertificateProviderPath
	WhatHappensNext                        CertificateProviderPath
	WhatIsYourHomeAddress                  CertificateProviderPath
	WhoIsEligible                          CertificateProviderPath
	YourPreferredLanguage                  CertificateProviderPath
	YourRole                               CertificateProviderPath
}

type HealthCheckPaths struct {
	Service    Path
	Dependency Path
}

type SupporterPaths struct {
	EnterOrganisationName Path
	EnterReferenceNumber  Path
	EnterYourName         Path
	InviteExpired         Path
	Login                 Path
	LoginCallback         Path
	OrganisationDeleted   Path
	SigningInAdvice       Path
	Start                 Path

	ConfirmDonorCanInteractOnline SupporterPath
	ContactOPGForPaperForms       SupporterPath
	Dashboard                     SupporterPath
	DeleteOrganisation            SupporterPath
	EditMember                    SupporterPath
	EditOrganisationName          SupporterPath
	InviteMember                  SupporterPath
	InviteMemberConfirmation      SupporterPath
	ManageTeamMembers             SupporterPath
	OrganisationCreated           SupporterPath
	OrganisationDetails           SupporterPath

	ViewLPA     SupporterLpaPath
	DonorAccess SupporterLpaPath
}

type AppPaths struct {
	Attorney            AttorneyPaths
	CertificateProvider CertificateProviderPaths
	Supporter           SupporterPaths
	HealthCheck         HealthCheckPaths

	AttorneyFixtures                   Path
	AuthRedirect                       Path
	CertificateProviderFixtures        Path
	CertificateProviderStart           Path
	CookiesConsent                     Path
	Dashboard                          Path
	DashboardFixtures                  Path
	DonorSubByLpaUID                   Path
	EnterAccessCode                    Path
	Fixtures                           Path
	Login                              Path
	LoginCallback                      Path
	LpaDeleted                         Path
	LpaWithdrawn                       Path
	Root                               Path
	SignOut                            Path
	Start                              Path
	SupporterFixtures                  Path
	YourLegalRightsAndResponsibilities Path

	AboutPayment                                         LpaPath
	AddCorrespondent                                     LpaPath
	AreYouApplyingForFeeDiscountOrExemption              LpaPath
	CertificateProviderAddress                           LpaPath
	CertificateProviderDetails                           LpaPath
	CertificateProviderOptOut                            LpaPath
	ChangeCertificateProviderMobileNumber                LpaPath
	ChangeIndependentWitnessMobileNumber                 LpaPath
	CheckYouCanSign                                      LpaPath
	CheckYourLpa                                         LpaPath
	ChooseAttorneys                                      LpaPath
	ChooseAttorneysAddress                               LpaPath
	ChooseAttorneysGuidance                              LpaPath
	ChooseAttorneysSummary                               LpaPath
	ChooseNewCertificateProvider                         LpaPath
	ChoosePeopleToNotify                                 LpaPath
	ChoosePeopleToNotifyAddress                          LpaPath
	ChoosePeopleToNotifySummary                          LpaPath
	ChooseReplacementAttorneys                           LpaPath
	ChooseReplacementAttorneysAddress                    LpaPath
	ChooseReplacementAttorneysSummary                    LpaPath
	ChooseYourCertificateProvider                        LpaPath
	ConfirmPersonAllowedToVouch                          LpaPath
	ConfirmYourCertificateProviderIsNotRelated           LpaPath
	DeleteThisLpa                                        LpaPath
	DoYouWantReplacementAttorneys                        LpaPath
	DoYouWantToNotifyPeople                              LpaPath
	EnterCorrespondentAddress                            LpaPath
	EnterCorrespondentDetails                            LpaPath
	EnterReplacementTrustCorporation                     LpaPath
	EnterReplacementTrustCorporationAddress              LpaPath
	EnterTrustCorporation                                LpaPath
	EnterTrustCorporationAddress                         LpaPath
	EnterVoucher                                         LpaPath
	EvidenceRequired                                     LpaPath
	EvidenceSuccessfullyUploaded                         LpaPath
	FeeApproved                                          LpaPath
	FeeDenied                                            LpaPath
	GettingHelpSigning                                   LpaPath
	HowDoYouKnowYourCertificateProvider                  LpaPath
	HowLongHaveYouKnownCertificateProvider               LpaPath
	HowShouldAttorneysMakeDecisions                      LpaPath
	HowShouldReplacementAttorneysMakeDecisions           LpaPath
	HowShouldReplacementAttorneysStepIn                  LpaPath
	HowToConfirmYourIdentityAndSign                      LpaPath
	HowToSendEvidence                                    LpaPath
	HowWouldCertificateProviderPreferToCarryOutTheirRole LpaPath
	HowWouldYouLikeToSendEvidence                        LpaPath
	IdentityConfirmed                                    LpaPath
	IdentityWithOneLogin                                 LpaPath
	IdentityWithOneLoginCallback                         LpaPath
	LifeSustainingTreatment                              LpaPath
	LpaDetailsSaved                                      LpaPath
	LpaType                                              LpaPath
	LpaYourLegalRightsAndResponsibilities                LpaPath
	MakeANewLPA                                          LpaPath
	NeedHelpSigningConfirmation                          LpaPath
	PaymentConfirmation                                  LpaPath
	PreviousApplicationNumber                            LpaPath
	PreviousFee                                          LpaPath
	Progress                                             LpaPath
	ProveYourIdentity                                    LpaPath
	ReadYourLpa                                          LpaPath
	RemoveAttorney                                       LpaPath
	RemovePersonToNotify                                 LpaPath
	RemoveReplacementAttorney                            LpaPath
	RemoveReplacementTrustCorporation                    LpaPath
	RemoveTrustCorporation                               LpaPath
	ResendCertificateProviderCode                        LpaPath
	ResendIndependentWitnessCode                         LpaPath
	Restrictions                                         LpaPath
	SendUsYourEvidenceByPost                             LpaPath
	SignTheLpaOnBehalf                                   LpaPath
	SignYourLpa                                          LpaPath
	TaskList                                             LpaPath
	UploadEvidence                                       LpaPath
	UploadEvidenceSSE                                    LpaPath
	UseExistingAddress                                   LpaPath
	WeHaveUpdatedYourDetails                             LpaPath
	WhatACertificateProviderDoes                         LpaPath
	WhatHappensNextPostEvidence                          LpaPath
	WhenCanTheLpaBeUsed                                  LpaPath
	WhichFeeTypeAreYouApplyingFor                        LpaPath
	WhoCanCorrespondentsDetailsBeSharedWith              LpaPath
	WithdrawThisLpa                                      LpaPath
	WitnessingAsCertificateProvider                      LpaPath
	WitnessingAsIndependentWitness                       LpaPath
	WitnessingYourSignature                              LpaPath
	YouCannotSignYourLpaYet                              LpaPath
	YouHaveSubmittedYourLpa                              LpaPath
	YourAddress                                          LpaPath
	YourAuthorisedSignatory                              LpaPath
	YourDateOfBirth                                      LpaPath
	YourDetails                                          LpaPath
	YourIndependentWitness                               LpaPath
	YourIndependentWitnessAddress                        LpaPath
	YourIndependentWitnessMobile                         LpaPath
	YourName                                             LpaPath
	YourPreferredLanguage                                LpaPath
}

var Paths = AppPaths{
	CertificateProvider: CertificateProviderPaths{
		CertificateProvided:                             "/certificate-provided",
		ConfirmDontWantToBeCertificateProvider:          "/confirm-you-do-not-want-to-be-a-certificate-provider",
		ConfirmDontWantToBeCertificateProviderLoggedOut: "/confirm-you-do-not-want-to-be-a-certificate-provider",
		ConfirmYourDetails:                              "/confirm-your-details",
		EnterDateOfBirth:                                "/enter-date-of-birth",
		EnterReferenceNumber:                            "/certificate-provider-enter-reference-number",
		EnterReferenceNumberOptOut:                      "/certificate-provider-enter-reference-number-opt-out",
		IdentityWithOneLogin:                            "/identity-with-one-login",
		IdentityWithOneLoginCallback:                    "/identity-with-one-login-callback",
		Login:                                           "/certificate-provider-login",
		LoginCallback:                                   "/certificate-provider-login-callback",
		ProveYourIdentity:                               "/prove-your-identity",
		ProvideCertificate:                              "/provide-certificate",
		ReadTheLpa:                                      "/read-the-lpa",
		TaskList:                                        "/task-list",
		WhatHappensNext:                                 "/what-happens-next",
		WhatIsYourHomeAddress:                           "/what-is-your-home-address",
		WhoIsEligible:                                   "/certificate-provider-who-is-eligible",
		YouHaveDecidedNotToBeACertificateProvider:       "/you-have-decided-not-to-be-a-certificate-provider",
		YourPreferredLanguage:                           "/your-preferred-language",
		YourRole:                                        "/your-role",
	},

	Attorney: AttorneyPaths{
		CodeOfConduct:             "/code-of-conduct",
		ConfirmYourDetails:        "/confirm-your-details",
		EnterReferenceNumber:      "/attorney-enter-reference-number",
		Login:                     "/attorney-login",
		LoginCallback:             "/attorney-login-callback",
		MobileNumber:              "/mobile-number",
		Progress:                  "/progress",
		ReadTheLpa:                "/read-the-lpa",
		RightsAndResponsibilities: "/legal-rights-and-responsibilities",
		Sign:                      "/sign",
		Start:                     "/attorney-start",
		TaskList:                  "/task-list",
		WhatHappensNext:           "/what-happens-next",
		WhatHappensWhenYouSign:    "/what-happens-when-you-sign-the-lpa",
		WouldLikeSecondSignatory:  "/would-like-second-signatory",
		YourPreferredLanguage:     "/your-preferred-language",
	},

	Supporter: SupporterPaths{
		EnterOrganisationName: "/enter-the-name-of-your-organisation-or-company",
		EnterReferenceNumber:  "/supporter-reference-number",
		EnterYourName:         "/enter-your-name",
		Login:                 "/supporter-login",
		LoginCallback:         "/supporter-login-callback",
		OrganisationDeleted:   "/organisation-deleted",
		SigningInAdvice:       "/signing-in-with-govuk-one-login",
		Start:                 "/supporter-start",

		ConfirmDonorCanInteractOnline: "/confirm-donor-can-interact-online",
		ContactOPGForPaperForms:       "/contact-opg-for-paper-forms",
		Dashboard:                     "/dashboard",
		DeleteOrganisation:            "/manage-organisation/organisation-details/delete-organisation",
		EditMember:                    "/manage-organisation/manage-team-members/edit-team-member",
		EditOrganisationName:          "/manage-organisation/organisation-details/edit-organisation-name",
		InviteExpired:                 "/invite-expired",
		InviteMember:                  "/invite-member",
		InviteMemberConfirmation:      "/invite-member-confirmation",
		ManageTeamMembers:             "/manage-organisation/manage-team-members",
		OrganisationCreated:           "/organisation-or-company-created",
		OrganisationDetails:           "/manage-organisation/organisation-details",
		ViewLPA:                       "/view-lpa",

		DonorAccess: "/donor-access",
	},

	HealthCheck: HealthCheckPaths{
		Service:    "/health-check/service",
		Dependency: "/health-check/dependency",
	},

	AboutPayment:                                         "/about-payment",
	AddCorrespondent:                                     "/add-correspondent",
	AreYouApplyingForFeeDiscountOrExemption:              "/are-you-applying-for-fee-discount-or-exemption",
	AttorneyFixtures:                                     "/fixtures/attorney",
	AuthRedirect:                                         "/auth/redirect",
	CertificateProviderAddress:                           "/certificate-provider-address",
	CertificateProviderDetails:                           "/certificate-provider-details",
	CertificateProviderFixtures:                          "/fixtures/certificate-provider",
	CertificateProviderOptOut:                            "/certificate-provider-opt-out",
	CertificateProviderStart:                             "/certificate-provider-start",
	ChangeCertificateProviderMobileNumber:                "/change-certificate-provider-mobile-number",
	ChangeIndependentWitnessMobileNumber:                 "/change-independent-witness-mobile-number",
	CheckYouCanSign:                                      "/check-you-can-sign",
	CheckYourLpa:                                         "/check-your-lpa",
	ChooseAttorneys:                                      "/choose-attorneys",
	ChooseAttorneysAddress:                               "/choose-attorneys-address",
	ChooseAttorneysGuidance:                              "/choose-attorneys-guidance",
	ChooseAttorneysSummary:                               "/choose-attorneys-summary",
	ChooseNewCertificateProvider:                         "/choose-new-certificate-provider",
	ChoosePeopleToNotify:                                 "/choose-people-to-notify",
	ChoosePeopleToNotifyAddress:                          "/choose-people-to-notify-address",
	ChoosePeopleToNotifySummary:                          "/choose-people-to-notify-summary",
	ChooseReplacementAttorneys:                           "/choose-replacement-attorneys",
	ChooseReplacementAttorneysAddress:                    "/choose-replacement-attorneys-address",
	ChooseReplacementAttorneysSummary:                    "/choose-replacement-attorneys-summary",
	ChooseYourCertificateProvider:                        "/choose-your-certificate-provider",
	ConfirmPersonAllowedToVouch:                          "/confirm-person-allowed-to-vouch",
	ConfirmYourCertificateProviderIsNotRelated:           "/confirm-your-certificate-provider-is-not-related",
	CookiesConsent:                                       "/cookies-consent",
	Dashboard:                                            "/dashboard",
	DashboardFixtures:                                    "/fixtures/dashboard",
	DeleteThisLpa:                                        "/delete-this-lpa",
	DoYouWantReplacementAttorneys:                        "/do-you-want-replacement-attorneys",
	DoYouWantToNotifyPeople:                              "/do-you-want-to-notify-people",
	EnterAccessCode:                                      "/enter-access-code",
	EnterCorrespondentAddress:                            "/enter-correspondent-address",
	EnterCorrespondentDetails:                            "/enter-correspondent-details",
	EnterReplacementTrustCorporation:                     "/enter-replacement-trust-corporation",
	EnterReplacementTrustCorporationAddress:              "/enter-replacement-trust-corporation-address",
	EnterTrustCorporation:                                "/enter-trust-corporation",
	EnterTrustCorporationAddress:                         "/enter-trust-corporation-address",
	EnterVoucher:                                         "/enter-voucher",
	EvidenceRequired:                                     "/evidence-required",
	EvidenceSuccessfullyUploaded:                         "/evidence-successfully-uploaded",
	FeeApproved:                                          "/fee-approved",
	FeeDenied:                                            "/fee-denied",
	Fixtures:                                             "/fixtures",
	GettingHelpSigning:                                   "/getting-help-signing",
	HowDoYouKnowYourCertificateProvider:                  "/how-do-you-know-your-certificate-provider",
	HowLongHaveYouKnownCertificateProvider:               "/how-long-have-you-known-certificate-provider",
	HowShouldAttorneysMakeDecisions:                      "/how-should-attorneys-make-decisions",
	HowShouldReplacementAttorneysMakeDecisions:           "/how-should-replacement-attorneys-make-decisions",
	HowShouldReplacementAttorneysStepIn:                  "/how-should-replacement-attorneys-step-in",
	HowToConfirmYourIdentityAndSign:                      "/how-to-confirm-your-identity-and-sign",
	HowToSendEvidence:                                    "/how-to-send-evidence",
	HowWouldCertificateProviderPreferToCarryOutTheirRole: "/how-would-certificate-provider-prefer-to-carry-out-their-role",
	HowWouldYouLikeToSendEvidence:                        "/how-would-you-like-to-send-evidence",
	IdentityConfirmed:                                    "/identity-confirmed",
	IdentityWithOneLogin:                                 "/id/one-login",
	IdentityWithOneLoginCallback:                         "/id/one-login/callback",
	LifeSustainingTreatment:                              "/life-sustaining-treatment",
	Login:                                                "/login",
	LoginCallback:                                        "/login-callback",
	LpaDeleted:                                           "/lpa-deleted",
	LpaDetailsSaved:                                      "/lpa-details-saved",
	LpaType:                                              "/lpa-type",
	LpaWithdrawn:                                         "/lpa-withdrawn",
	LpaYourLegalRightsAndResponsibilities:                "/your-legal-rights-and-responsibilities",
	MakeANewLPA:                                          "/make-a-new-lpa",
	NeedHelpSigningConfirmation:                          "/need-help-signing-confirmation",
	PaymentConfirmation:                                  "/payment-confirmation",
	PreviousApplicationNumber:                            "/previous-application-number",
	PreviousFee:                                          "/how-much-did-you-previously-pay-for-your-lpa",
	Progress:                                             "/progress",
	ProveYourIdentity:                                    "/prove-your-identity",
	ReadYourLpa:                                          "/read-your-lpa",
	RemoveAttorney:                                       "/remove-attorney",
	RemovePersonToNotify:                                 "/remove-person-to-notify",
	RemoveReplacementAttorney:                            "/remove-replacement-attorney",
	RemoveReplacementTrustCorporation:                    "/remove-replacement-trust-corporation",
	RemoveTrustCorporation:                               "/remove-trust-corporation",
	ResendCertificateProviderCode:                        "/resend-certificate-provider-code",
	ResendIndependentWitnessCode:                         "/resend-independent-witness-code",
	Restrictions:                                         "/restrictions",
	Root:                                                 "/",
	SendUsYourEvidenceByPost:                             "/send-us-your-evidence-by-post",
	SignOut:                                              "/sign-out",
	SignTheLpaOnBehalf:                                   "/sign-the-lpa-on-behalf",
	SignYourLpa:                                          "/sign-your-lpa",
	Start:                                                "/start",
	SupporterFixtures:                                    "/fixtures/supporter",
	TaskList:                                             "/task-list",
	UploadEvidence:                                       "/upload-evidence",
	UploadEvidenceSSE:                                    "/upload-evidence-sse",
	UseExistingAddress:                                   "/use-existing-address",
	WeHaveUpdatedYourDetails:                             "/we-have-updated-your-details",
	WhatACertificateProviderDoes:                         "/what-a-certificate-provider-does",
	WhatHappensNextPostEvidence:                          "/what-happens-next-post-evidence",
	WhenCanTheLpaBeUsed:                                  "/when-can-the-lpa-be-used",
	WhichFeeTypeAreYouApplyingFor:                        "/which-fee-type-are-you-applying-for",
	WhoCanCorrespondentsDetailsBeSharedWith:              "/who-can-correspondents-details-be-shared-with",
	WithdrawThisLpa:                                      "/withdraw-this-lpa",
	WitnessingAsCertificateProvider:                      "/witnessing-as-certificate-provider",
	WitnessingAsIndependentWitness:                       "/witnessing-as-independent-witness",
	WitnessingYourSignature:                              "/witnessing-your-signature",
	YouCannotSignYourLpaYet:                              "/you-cannot-sign-your-lpa-yet",
	YouHaveSubmittedYourLpa:                              "/you-have-submitted-your-lpa",
	YourAddress:                                          "/your-address",
	YourAuthorisedSignatory:                              "/your-authorised-signatory",
	YourDateOfBirth:                                      "/your-date-of-birth",
	YourDetails:                                          "/your-details",
	YourIndependentWitness:                               "/your-independent-witness",
	YourIndependentWitnessAddress:                        "/your-independent-witness-address",
	YourIndependentWitnessMobile:                         "/your-independent-witness-mobile",
	YourLegalRightsAndResponsibilities:                   "/your-legal-rights-and-responsibilities",
	YourName:                                             "/your-name",
	YourPreferredLanguage:                                "/your-preferred-language",
}

func DonorCanGoTo(donor *actor.DonorProvidedDetails, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/lpa/") {
		_, lpaPath, _ := strings.Cut(strings.TrimPrefix(path, "/lpa/"), "/")
		return LpaPath("/" + lpaPath).canVisit(donor)
	}

	return true
}

func CertificateProviderCanGoTo(certificateProvider *actor.CertificateProviderProvidedDetails, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/certificate-provider/") {
		_, certificateProviderPath, _ := strings.Cut(strings.TrimPrefix(path, "/certificate-provider/"), "/")
		return CertificateProviderPath("/" + certificateProviderPath).canVisit(certificateProvider)
	}

	return true
}

func AttorneyCanGoTo(attorney *actor.AttorneyProvidedDetails, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/attorney/") {
		_, attorneyPath, _ := strings.Cut(strings.TrimPrefix(path, "/attorney/"), "/")
		return AttorneyPath("/" + attorneyPath).canVisit(attorney)
	}

	return true
}
