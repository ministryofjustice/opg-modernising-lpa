package page

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
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

type AttorneyPaths struct {
	ConfirmDontWantToBeAttorneyLoggedOut Path
	EnterReferenceNumber                 Path
	EnterReferenceNumberOptOut           Path
	Login                                Path
	LoginCallback                        Path
	Start                                Path
	YouHaveDecidedNotToBeAttorney        Path

	CodeOfConduct               attorney.Path
	ConfirmDontWantToBeAttorney attorney.Path
	ConfirmYourDetails          attorney.Path
	MobileNumber                attorney.Path
	Progress                    attorney.Path
	ReadTheLpa                  attorney.Path
	RightsAndResponsibilities   attorney.Path
	Sign                        attorney.Path
	TaskList                    attorney.Path
	WhatHappensNext             attorney.Path
	WhatHappensWhenYouSign      attorney.Path
	WouldLikeSecondSignatory    attorney.Path
	YourPreferredLanguage       attorney.Path
}

type CertificateProviderPaths struct {
	Login                                           Path
	LoginCallback                                   Path
	EnterReferenceNumber                            Path
	EnterReferenceNumberOptOut                      Path
	ConfirmDontWantToBeCertificateProviderLoggedOut Path
	YouHaveDecidedNotToBeCertificateProvider        Path

	CertificateProvided                    certificateprovider.Path
	ConfirmDontWantToBeCertificateProvider certificateprovider.Path
	ConfirmYourDetails                     certificateprovider.Path
	EnterDateOfBirth                       certificateprovider.Path
	IdentityWithOneLogin                   certificateprovider.Path
	IdentityWithOneLoginCallback           certificateprovider.Path
	OneLoginIdentityDetails                certificateprovider.Path
	ProveYourIdentity                      certificateprovider.Path
	ProvideCertificate                     certificateprovider.Path
	ReadTheLpa                             certificateprovider.Path
	TaskList                               certificateprovider.Path
	UnableToConfirmIdentity                certificateprovider.Path
	WhatHappensNext                        certificateprovider.Path
	WhatIsYourHomeAddress                  certificateprovider.Path
	WhoIsEligible                          certificateprovider.Path
	YourPreferredLanguage                  certificateprovider.Path
	YourRole                               certificateprovider.Path
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

	ConfirmDonorCanInteractOnline supporter.Path
	ContactOPGForPaperForms       supporter.Path
	Dashboard                     supporter.Path
	DeleteOrganisation            supporter.Path
	EditMember                    supporter.Path
	EditOrganisationName          supporter.Path
	InviteMember                  supporter.Path
	InviteMemberConfirmation      supporter.Path
	ManageTeamMembers             supporter.Path
	OrganisationCreated           supporter.Path
	OrganisationDetails           supporter.Path

	ViewLPA     supporter.LpaPath
	DonorAccess supporter.LpaPath
}

type AppPaths struct {
	Attorney            AttorneyPaths
	CertificateProvider CertificateProviderPaths
	Supporter           SupporterPaths
	HealthCheck         HealthCheckPaths

	AttorneyFixtures            Path
	AuthRedirect                Path
	CertificateProviderFixtures Path
	CertificateProviderStart    Path
	CookiesConsent              Path
	Dashboard                   Path
	DashboardFixtures           Path
	EnterAccessCode             Path
	Fixtures                    Path
	Login                       Path
	LoginCallback               Path
	LpaDeleted                  Path
	LpaWithdrawn                Path
	Root                        Path
	SignOut                     Path
	Start                       Path
	SupporterFixtures           Path

	AboutPayment                                         donor.Path
	AddCorrespondent                                     donor.Path
	AreYouApplyingForFeeDiscountOrExemption              donor.Path
	BecauseYouHaveChosenJointly                          donor.Path
	BecauseYouHaveChosenJointlyForSomeSeverallyForOthers donor.Path
	CanYouSignYourLpa                                    donor.Path
	CertificateProviderAddress                           donor.Path
	CertificateProviderDetails                           donor.Path
	CertificateProviderOptOut                            donor.Path
	ChangeCertificateProviderMobileNumber                donor.Path
	ChangeIndependentWitnessMobileNumber                 donor.Path
	CheckYouCanSign                                      donor.Path
	CheckYourDetails                                     donor.Path
	CheckYourLpa                                         donor.Path
	ChooseAttorneys                                      donor.Path
	ChooseAttorneysAddress                               donor.Path
	ChooseAttorneysGuidance                              donor.Path
	ChooseAttorneysSummary                               donor.Path
	ChooseNewCertificateProvider                         donor.Path
	ChoosePeopleToNotify                                 donor.Path
	ChoosePeopleToNotifyAddress                          donor.Path
	ChoosePeopleToNotifySummary                          donor.Path
	ChooseReplacementAttorneys                           donor.Path
	ChooseReplacementAttorneysAddress                    donor.Path
	ChooseReplacementAttorneysSummary                    donor.Path
	ChooseYourCertificateProvider                        donor.Path
	ConfirmPersonAllowedToVouch                          donor.Path
	ConfirmYourCertificateProviderIsNotRelated           donor.Path
	DeleteThisLpa                                        donor.Path
	DoYouWantReplacementAttorneys                        donor.Path
	DoYouWantToNotifyPeople                              donor.Path
	EnterCorrespondentAddress                            donor.Path
	EnterCorrespondentDetails                            donor.Path
	EnterReplacementTrustCorporation                     donor.Path
	EnterReplacementTrustCorporationAddress              donor.Path
	EnterTrustCorporation                                donor.Path
	EnterTrustCorporationAddress                         donor.Path
	EnterVoucher                                         donor.Path
	EvidenceRequired                                     donor.Path
	EvidenceSuccessfullyUploaded                         donor.Path
	FeeApproved                                          donor.Path
	FeeDenied                                            donor.Path
	GettingHelpSigning                                   donor.Path
	HowDoYouKnowYourCertificateProvider                  donor.Path
	HowLongHaveYouKnownCertificateProvider               donor.Path
	HowShouldAttorneysMakeDecisions                      donor.Path
	HowShouldReplacementAttorneysMakeDecisions           donor.Path
	HowShouldReplacementAttorneysStepIn                  donor.Path
	HowToConfirmYourIdentityAndSign                      donor.Path
	HowToSendEvidence                                    donor.Path
	HowWouldCertificateProviderPreferToCarryOutTheirRole donor.Path
	HowWouldYouLikeToSendEvidence                        donor.Path
	IdentityWithOneLogin                                 donor.Path
	IdentityWithOneLoginCallback                         donor.Path
	LifeSustainingTreatment                              donor.Path
	LpaDetailsSaved                                      donor.Path
	LpaType                                              donor.Path
	LpaYourLegalRightsAndResponsibilities                donor.Path
	MakeANewLPA                                          donor.Path
	NeedHelpSigningConfirmation                          donor.Path
	OneLoginIdentityDetails                              donor.Path
	PaymentConfirmation                                  donor.Path
	PreviousApplicationNumber                            donor.Path
	PreviousFee                                          donor.Path
	Progress                                             donor.Path
	ProveYourIdentity                                    donor.Path
	ReadYourLpa                                          donor.Path
	RegisterWithCourtOfProtection                        donor.Path
	RemoveAttorney                                       donor.Path
	RemovePersonToNotify                                 donor.Path
	RemoveReplacementAttorney                            donor.Path
	RemoveReplacementTrustCorporation                    donor.Path
	RemoveTrustCorporation                               donor.Path
	ResendCertificateProviderCode                        donor.Path
	ResendIndependentWitnessCode                         donor.Path
	Restrictions                                         donor.Path
	SendUsYourEvidenceByPost                             donor.Path
	SignTheLpaOnBehalf                                   donor.Path
	SignYourLpa                                          donor.Path
	TaskList                                             donor.Path
	UnableToConfirmIdentity                              donor.Path
	UploadEvidence                                       donor.Path
	UploadEvidenceSSE                                    donor.Path
	UseExistingAddress                                   donor.Path
	WeHaveContactedVoucher                               donor.Path
	WeHaveReceivedVoucherDetails                         donor.Path
	WeHaveUpdatedYourDetails                             donor.Path
	WhatACertificateProviderDoes                         donor.Path
	WhatHappensNextPostEvidence                          donor.Path
	WhatHappensNextRegisteringWithCourtOfProtection      donor.Path
	WhatIsVouching                                       donor.Path
	WhatYouCanDoNow                                      donor.Path
	WhenCanTheLpaBeUsed                                  donor.Path
	WhichFeeTypeAreYouApplyingFor                        donor.Path
	WithdrawThisLpa                                      donor.Path
	WitnessingAsCertificateProvider                      donor.Path
	WitnessingAsIndependentWitness                       donor.Path
	WitnessingYourSignature                              donor.Path
	YouCannotSignYourLpaYet                              donor.Path
	YouHaveSubmittedYourLpa                              donor.Path
	YourAddress                                          donor.Path
	YourAuthorisedSignatory                              donor.Path
	YourDateOfBirth                                      donor.Path
	YourDetails                                          donor.Path
	YourEmail                                            donor.Path
	YourIndependentWitness                               donor.Path
	YourIndependentWitnessAddress                        donor.Path
	YourIndependentWitnessMobile                         donor.Path
	YourLegalRightsAndResponsibilitiesIfYouMakeLpa       donor.Path
	YourLpaLanguage                                      donor.Path
	YourName                                             donor.Path
	YourPreferredLanguage                                donor.Path
}

var Paths = AppPaths{
	CertificateProvider: CertificateProviderPaths{
		ConfirmDontWantToBeCertificateProviderLoggedOut: "/confirm-you-do-not-want-to-be-a-certificate-provider",
		EnterReferenceNumber:                            "/certificate-provider-enter-reference-number",
		EnterReferenceNumberOptOut:                      "/certificate-provider-enter-reference-number-opt-out",
		Login:                                           "/certificate-provider-login",
		LoginCallback:                                   "/certificate-provider-login-callback",
		YouHaveDecidedNotToBeCertificateProvider:        "/you-have-decided-not-to-be-a-certificate-provider",

		CertificateProvided:                    certificateprovider.PathCertificateProvided,
		ConfirmDontWantToBeCertificateProvider: certificateprovider.PathConfirmDontWantToBeCertificateProvider,
		ConfirmYourDetails:                     certificateprovider.PathConfirmYourDetails,
		EnterDateOfBirth:                       certificateprovider.PathEnterDateOfBirth,
		IdentityWithOneLogin:                   certificateprovider.PathIdentityWithOneLogin,
		IdentityWithOneLoginCallback:           certificateprovider.PathIdentityWithOneLoginCallback,
		ProveYourIdentity:                      certificateprovider.PathProveYourIdentity,
		OneLoginIdentityDetails:                certificateprovider.PathOneLoginIdentityDetails,
		ProvideCertificate:                     certificateprovider.PathProvideCertificate,
		ReadTheLpa:                             certificateprovider.PathReadTheLpa,
		TaskList:                               certificateprovider.PathTaskList,
		UnableToConfirmIdentity:                certificateprovider.PathUnableToConfirmIdentity,
		WhatHappensNext:                        certificateprovider.PathWhatHappensNext,
		WhatIsYourHomeAddress:                  certificateprovider.PathWhatIsYourHomeAddress,
		WhoIsEligible:                          certificateprovider.PathWhoIsEligible,
		YourPreferredLanguage:                  certificateprovider.PathYourPreferredLanguage,
		YourRole:                               certificateprovider.PathYourRole,
	},

	Attorney: AttorneyPaths{
		ConfirmDontWantToBeAttorneyLoggedOut: "/confirm-you-do-not-want-to-be-an-attorney",
		EnterReferenceNumber:                 "/attorney-enter-reference-number",
		EnterReferenceNumberOptOut:           "/attorney-enter-reference-number-opt-out",
		Login:                                "/attorney-login",
		LoginCallback:                        "/attorney-login-callback",
		Start:                                "/attorney-start",
		YouHaveDecidedNotToBeAttorney:        "/you-have-decided-not-to-be-an-attorney",

		CodeOfConduct:               attorney.PathCodeOfConduct,
		ConfirmDontWantToBeAttorney: attorney.PathConfirmDontWantToBeAttorney,
		ConfirmYourDetails:          attorney.PathConfirmYourDetails,
		MobileNumber:                attorney.PathMobileNumber,
		Progress:                    attorney.PathProgress,
		ReadTheLpa:                  attorney.PathReadTheLpa,
		RightsAndResponsibilities:   attorney.PathRightsAndResponsibilities,
		Sign:                        attorney.PathSign,
		TaskList:                    attorney.PathTaskList,
		WhatHappensNext:             attorney.PathWhatHappensNext,
		WhatHappensWhenYouSign:      attorney.PathWhatHappensWhenYouSign,
		WouldLikeSecondSignatory:    attorney.PathWouldLikeSecondSignatory,
		YourPreferredLanguage:       attorney.PathYourPreferredLanguage,
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
		InviteExpired:         "/invite-expired",

		ConfirmDonorCanInteractOnline: supporter.PathConfirmDonorCanInteractOnline,
		ContactOPGForPaperForms:       supporter.PathContactOPGForPaperForms,
		Dashboard:                     supporter.PathDashboard,
		DeleteOrganisation:            supporter.PathDeleteOrganisation,
		EditMember:                    supporter.PathEditMember,
		EditOrganisationName:          supporter.PathEditOrganisationName,
		InviteMember:                  supporter.PathInviteMember,
		InviteMemberConfirmation:      supporter.PathInviteMemberConfirmation,
		ManageTeamMembers:             supporter.PathManageTeamMembers,
		OrganisationCreated:           supporter.PathOrganisationCreated,
		OrganisationDetails:           supporter.PathOrganisationDetails,
		ViewLPA:                       supporter.PathViewLPA,
		DonorAccess:                   supporter.PathDonorAccess,
	},

	HealthCheck: HealthCheckPaths{
		Service:    "/health-check/service",
		Dependency: "/health-check/dependency",
	},

	AttorneyFixtures:            "/fixtures/attorney",
	AuthRedirect:                "/auth/redirect",
	CertificateProviderFixtures: "/fixtures/certificate-provider",
	CertificateProviderStart:    "/certificate-provider-start",
	CookiesConsent:              "/cookies-consent",
	Dashboard:                   "/dashboard",
	DashboardFixtures:           "/fixtures/dashboard",
	EnterAccessCode:             "/enter-access-code",
	Fixtures:                    "/fixtures",
	Login:                       "/login",
	LoginCallback:               "/login-callback",
	LpaDeleted:                  "/lpa-deleted",
	LpaWithdrawn:                "/lpa-withdrawn",
	Root:                        "/",
	SignOut:                     "/sign-out",
	Start:                       "/start",
	SupporterFixtures:           "/fixtures/supporter",

	AboutPayment:                                         donor.PathAboutPayment,
	AddCorrespondent:                                     donor.PathAddCorrespondent,
	AreYouApplyingForFeeDiscountOrExemption:              donor.PathAreYouApplyingForFeeDiscountOrExemption,
	BecauseYouHaveChosenJointly:                          donor.PathBecauseYouHaveChosenJointly,
	BecauseYouHaveChosenJointlyForSomeSeverallyForOthers: donor.PathBecauseYouHaveChosenJointlyForSomeSeverallyForOthers,
	CanYouSignYourLpa:                                    donor.PathCanYouSignYourLpa,
	CertificateProviderAddress:                           donor.PathCertificateProviderAddress,
	CertificateProviderDetails:                           donor.PathCertificateProviderDetails,
	CertificateProviderOptOut:                            donor.PathCertificateProviderOptOut,
	ChangeCertificateProviderMobileNumber:                donor.PathChangeCertificateProviderMobileNumber,
	ChangeIndependentWitnessMobileNumber:                 donor.PathChangeIndependentWitnessMobileNumber,
	CheckYouCanSign:                                      donor.PathCheckYouCanSign,
	CheckYourDetails:                                     donor.PathCheckYourDetails,
	CheckYourLpa:                                         donor.PathCheckYourLpa,
	ChooseAttorneys:                                      donor.PathChooseAttorneys,
	ChooseAttorneysAddress:                               donor.PathChooseAttorneysAddress,
	ChooseAttorneysGuidance:                              donor.PathChooseAttorneysGuidance,
	ChooseAttorneysSummary:                               donor.PathChooseAttorneysSummary,
	ChooseNewCertificateProvider:                         donor.PathChooseNewCertificateProvider,
	ChoosePeopleToNotify:                                 donor.PathChoosePeopleToNotify,
	ChoosePeopleToNotifyAddress:                          donor.PathChoosePeopleToNotifyAddress,
	ChoosePeopleToNotifySummary:                          donor.PathChoosePeopleToNotifySummary,
	ChooseReplacementAttorneys:                           donor.PathChooseReplacementAttorneys,
	ChooseReplacementAttorneysAddress:                    donor.PathChooseReplacementAttorneysAddress,
	ChooseReplacementAttorneysSummary:                    donor.PathChooseReplacementAttorneysSummary,
	ChooseYourCertificateProvider:                        donor.PathChooseYourCertificateProvider,
	ConfirmPersonAllowedToVouch:                          donor.PathConfirmPersonAllowedToVouch,
	ConfirmYourCertificateProviderIsNotRelated:           donor.PathConfirmYourCertificateProviderIsNotRelated,
	DeleteThisLpa:                                        donor.PathDeleteThisLpa,
	DoYouWantReplacementAttorneys:                        donor.PathDoYouWantReplacementAttorneys,
	DoYouWantToNotifyPeople:                              donor.PathDoYouWantToNotifyPeople,
	EnterCorrespondentAddress:                            donor.PathEnterCorrespondentAddress,
	EnterCorrespondentDetails:                            donor.PathEnterCorrespondentDetails,
	EnterReplacementTrustCorporation:                     donor.PathEnterReplacementTrustCorporation,
	EnterReplacementTrustCorporationAddress:              donor.PathEnterReplacementTrustCorporationAddress,
	EnterTrustCorporation:                                donor.PathEnterTrustCorporation,
	EnterTrustCorporationAddress:                         donor.PathEnterTrustCorporationAddress,
	EnterVoucher:                                         donor.PathEnterVoucher,
	EvidenceRequired:                                     donor.PathEvidenceRequired,
	EvidenceSuccessfullyUploaded:                         donor.PathEvidenceSuccessfullyUploaded,
	FeeApproved:                                          donor.PathFeeApproved,
	FeeDenied:                                            donor.PathFeeDenied,
	GettingHelpSigning:                                   donor.PathGettingHelpSigning,
	HowDoYouKnowYourCertificateProvider:                  donor.PathHowDoYouKnowYourCertificateProvider,
	HowLongHaveYouKnownCertificateProvider:               donor.PathHowLongHaveYouKnownCertificateProvider,
	HowShouldAttorneysMakeDecisions:                      donor.PathHowShouldAttorneysMakeDecisions,
	HowShouldReplacementAttorneysMakeDecisions:           donor.PathHowShouldReplacementAttorneysMakeDecisions,
	HowShouldReplacementAttorneysStepIn:                  donor.PathHowShouldReplacementAttorneysStepIn,
	HowToConfirmYourIdentityAndSign:                      donor.PathHowToConfirmYourIdentityAndSign,
	HowToSendEvidence:                                    donor.PathHowToSendEvidence,
	HowWouldCertificateProviderPreferToCarryOutTheirRole: donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole,
	HowWouldYouLikeToSendEvidence:                        donor.PathHowWouldYouLikeToSendEvidence,
	IdentityWithOneLogin:                                 donor.PathIdentityWithOneLogin,
	IdentityWithOneLoginCallback:                         donor.PathIdentityWithOneLoginCallback,
	LifeSustainingTreatment:                              donor.PathLifeSustainingTreatment,
	LpaDetailsSaved:                                      donor.PathLpaDetailsSaved,
	LpaType:                                              donor.PathLpaType,
	LpaYourLegalRightsAndResponsibilities:                donor.PathLpaYourLegalRightsAndResponsibilities,
	MakeANewLPA:                                          donor.PathMakeANewLPA,
	NeedHelpSigningConfirmation:                          donor.PathNeedHelpSigningConfirmation,
	OneLoginIdentityDetails:                              donor.PathOneLoginIdentityDetails,
	PaymentConfirmation:                                  donor.PathPaymentConfirmation,
	PreviousApplicationNumber:                            donor.PathPreviousApplicationNumber,
	PreviousFee:                                          donor.PathPreviousFee,
	Progress:                                             donor.PathProgress,
	ProveYourIdentity:                                    donor.PathProveYourIdentity,
	ReadYourLpa:                                          donor.PathReadYourLpa,
	RegisterWithCourtOfProtection:                        donor.PathRegisterWithCourtOfProtection,
	RemoveAttorney:                                       donor.PathRemoveAttorney,
	RemovePersonToNotify:                                 donor.PathRemovePersonToNotify,
	RemoveReplacementAttorney:                            donor.PathRemoveReplacementAttorney,
	RemoveReplacementTrustCorporation:                    donor.PathRemoveReplacementTrustCorporation,
	RemoveTrustCorporation:                               donor.PathRemoveTrustCorporation,
	ResendCertificateProviderCode:                        donor.PathResendCertificateProviderCode,
	ResendIndependentWitnessCode:                         donor.PathResendIndependentWitnessCode,
	Restrictions:                                         donor.PathRestrictions,
	SendUsYourEvidenceByPost:                             donor.PathSendUsYourEvidenceByPost,
	SignTheLpaOnBehalf:                                   donor.PathSignTheLpaOnBehalf,
	SignYourLpa:                                          donor.PathSignYourLpa,
	TaskList:                                             donor.PathTaskList,
	UnableToConfirmIdentity:                              donor.PathUnableToConfirmIdentity,
	UploadEvidence:                                       donor.PathUploadEvidence,
	UploadEvidenceSSE:                                    donor.PathUploadEvidenceSSE,
	UseExistingAddress:                                   donor.PathUseExistingAddress,
	WeHaveContactedVoucher:                               donor.PathWeHaveContactedVoucher,
	WeHaveReceivedVoucherDetails:                         donor.PathWeHaveReceivedVoucherDetails,
	WeHaveUpdatedYourDetails:                             donor.PathWeHaveUpdatedYourDetails,
	WhatACertificateProviderDoes:                         donor.PathWhatACertificateProviderDoes,
	WhatHappensNextPostEvidence:                          donor.PathWhatHappensNextPostEvidence,
	WhatHappensNextRegisteringWithCourtOfProtection:      donor.PathWhatHappensNextRegisteringWithCourtOfProtection,
	WhatIsVouching:                                       donor.PathWhatIsVouching,
	WhatYouCanDoNow:                                      donor.PathWhatYouCanDoNow,
	WhenCanTheLpaBeUsed:                                  donor.PathWhenCanTheLpaBeUsed,
	WhichFeeTypeAreYouApplyingFor:                        donor.PathWhichFeeTypeAreYouApplyingFor,
	WithdrawThisLpa:                                      donor.PathWithdrawThisLpa,
	WitnessingAsCertificateProvider:                      donor.PathWitnessingAsCertificateProvider,
	WitnessingAsIndependentWitness:                       donor.PathWitnessingAsIndependentWitness,
	WitnessingYourSignature:                              donor.PathWitnessingYourSignature,
	YouCannotSignYourLpaYet:                              donor.PathYouCannotSignYourLpaYet,
	YouHaveSubmittedYourLpa:                              donor.PathYouHaveSubmittedYourLpa,
	YourAddress:                                          donor.PathYourAddress,
	YourAuthorisedSignatory:                              donor.PathYourAuthorisedSignatory,
	YourDateOfBirth:                                      donor.PathYourDateOfBirth,
	YourDetails:                                          donor.PathYourDetails,
	YourEmail:                                            donor.PathYourEmail,
	YourIndependentWitness:                               donor.PathYourIndependentWitness,
	YourIndependentWitnessAddress:                        donor.PathYourIndependentWitnessAddress,
	YourIndependentWitnessMobile:                         donor.PathYourIndependentWitnessMobile,
	YourLegalRightsAndResponsibilitiesIfYouMakeLpa:       donor.PathYourLegalRightsAndResponsibilitiesIfYouMakeLpa,
	YourLpaLanguage:                                      donor.PathYourLpaLanguage,
	YourName:                                             donor.PathYourName,
	YourPreferredLanguage:                                donor.PathYourPreferredLanguage,
}
