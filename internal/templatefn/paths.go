package templatefn

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

type attorneyPaths struct {
	ConfirmDontWantToBeAttorneyLoggedOut page.Path
	EnterReferenceNumber                 page.Path
	EnterReferenceNumberOptOut           page.Path
	Login                                page.Path
	LoginCallback                        page.Path
	Start                                page.Path
	YouHaveDecidedNotToBeAttorney        page.Path

	CodeOfConduct               attorney.Path
	ConfirmDontWantToBeAttorney attorney.Path
	ConfirmYourDetails          attorney.Path
	PhoneNumber                 attorney.Path
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

type certificateProviderPaths struct {
	Login                                           page.Path
	LoginCallback                                   page.Path
	EnterReferenceNumber                            page.Path
	EnterReferenceNumberOptOut                      page.Path
	ConfirmDontWantToBeCertificateProviderLoggedOut page.Path
	YouHaveAlreadyProvidedACertificate              page.Path
	YouHaveAlreadyProvidedACertificateLoggedIn      page.Path
	YouHaveDecidedNotToBeCertificateProvider        page.Path

	CertificateProvided                    certificateprovider.Path
	ConfirmDontWantToBeCertificateProvider certificateprovider.Path
	ConfirmYourDetails                     certificateprovider.Path
	EnterDateOfBirth                       certificateprovider.Path
	IdentityWithOneLogin                   certificateprovider.Path
	IdentityWithOneLoginCallback           certificateprovider.Path
	OneLoginIdentityDetails                certificateprovider.Path
	ProveYourIdentity                      certificateprovider.Path
	ProvideCertificate                     certificateprovider.Path
	ReadTheDraftLpa                        certificateprovider.Path
	ReadTheLpa                             certificateprovider.Path
	TaskList                               certificateprovider.Path
	WhatHappensNext                        certificateprovider.Path
	WhatIsYourHomeAddress                  certificateprovider.Path
	WhoIsEligible                          certificateprovider.Path
	YourPreferredLanguage                  certificateprovider.Path
	YourRole                               certificateprovider.Path
}

type healthCheckPaths struct {
	Service    page.Path
	Dependency page.Path
}

type supporterPaths struct {
	EnterOrganisationName page.Path
	EnterReferenceNumber  page.Path
	EnterYourName         page.Path
	InviteExpired         page.Path
	Login                 page.Path
	LoginCallback         page.Path
	OrganisationDeleted   page.Path
	SigningInAdvice       page.Path
	Start                 page.Path

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

type voucherPaths struct {
	EnterReferenceNumber page.Path
	Login                page.Path
	Start                page.Path

	TaskList             voucher.Path
	YourName             voucher.Path
	IdentityWithOneLogin voucher.Path
}

type appPaths struct {
	Attorney            attorneyPaths
	CertificateProvider certificateProviderPaths
	Supporter           supporterPaths
	Voucher             voucherPaths
	HealthCheck         healthCheckPaths

	AccessibilityStatement      page.Path
	AttorneyFixtures            page.Path
	AddAnLPA                    page.Path
	AuthRedirect                page.Path
	CertificateProviderFixtures page.Path
	CertificateProviderStart    page.Path
	CookiesConsent              page.Path
	Dashboard                   page.Path
	DashboardFixtures           page.Path
	EnterAccessCode             page.Path
	Fixtures                    page.Path
	Login                       page.Path
	LoginCallback               page.Path
	LpaDeleted                  page.Path
	LpaWithdrawn                page.Path
	MakeOrAddAnLPA              page.Path
	Root                        page.Path
	SignOut                     page.Path
	Start                       page.Path
	SupporterFixtures           page.Path
	VoucherFixtures             page.Path
	VoucherStart                page.Path

	AddingRestrictionsAndConditions          page.Path
	ContactTheOfficeOfThePublicGuardian      page.Path
	Glossary                                 page.Path
	HowDecisionsAreMadeWithMultipleAttorneys page.Path
	HowToMakeAndRegisterYourLPA              page.Path
	HowToSelectAttorneysForAnLPA             page.Path
	ReplacementAttorneys                     page.Path
	TheTwoTypesOfLPA                         page.Path
	UnderstandingLifeSustainingTreatment     page.Path
	UnderstandingMentalCapacity              page.Path

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
	ChangeDonorEmail                                     donor.Path
	ChangeDonorMobileNumber                              donor.Path
	ChangeIndependentWitnessMobileNumber                 donor.Path
	CheckYouCanSign                                      donor.Path
	CheckYourDetails                                     donor.Path
	CheckYourLpa                                         donor.Path
	ChooseAttorneys                                      donor.Path
	ChooseAttorneysAddress                               donor.Path
	ChooseAttorneysGuidance                              donor.Path
	ChooseAttorneysSummary                               donor.Path
	ChooseCertificateProvider                            donor.Path
	ChooseNewCertificateProvider                         donor.Path
	ChoosePeopleToNotify                                 donor.Path
	ChoosePeopleToNotifyAddress                          donor.Path
	ChoosePeopleToNotifySummary                          donor.Path
	ChooseReplacementAttorneys                           donor.Path
	ChooseReplacementAttorneysAddress                    donor.Path
	ChooseReplacementAttorneysSummary                    donor.Path
	ChooseReplacementTrustCorporation                    donor.Path
	ChooseSomeoneToVouchForYou                           donor.Path
	ChooseTrustCorporation                               donor.Path
	ChooseYourCertificateProvider                        donor.Path
	ConfirmPersonAllowedToVouch                          donor.Path
	ConfirmYourCertificateProviderIsNotRelated           donor.Path
	DeleteThisLpa                                        donor.Path
	DoYouLiveInTheUK                                     donor.Path
	DoYouWantReplacementAttorneys                        donor.Path
	DoYouWantToNotifyPeople                              donor.Path
	EnterAttorney                                        donor.Path
	EnterCorrespondentAddress                            donor.Path
	EnterCorrespondentDetails                            donor.Path
	EnterReplacementAttorney                             donor.Path
	EnterReplacementTrustCorporation                     donor.Path
	EnterReplacementTrustCorporationAddress              donor.Path
	EnterTrustCorporation                                donor.Path
	EnterTrustCorporationAddress                         donor.Path
	EnterVoucher                                         donor.Path
	EvidenceRequired                                     donor.Path
	EvidenceSuccessfullyUploaded                         donor.Path
	GettingHelpSigning                                   donor.Path
	HowDoYouKnowYourCertificateProvider                  donor.Path
	HowLongHaveYouKnownCertificateProvider               donor.Path
	HowShouldAttorneysMakeDecisions                      donor.Path
	HowShouldReplacementAttorneysMakeDecisions           donor.Path
	HowShouldReplacementAttorneysStepIn                  donor.Path
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
	PayFee                                               donor.Path
	PaymentConfirmation                                  donor.Path
	PreviousApplicationNumber                            donor.Path
	PreviousFee                                          donor.Path
	Progress                                             donor.Path
	ProveYourIdentity                                    donor.Path
	ReadYourLpa                                          donor.Path
	ReceivingUpdatesAboutYourLpa                         donor.Path
	RegisterWithCourtOfProtection                        donor.Path
	RemoveAttorney                                       donor.Path
	RemoveCertificateProvider                            donor.Path
	RemoveCorrespondent                                  donor.Path
	RemovePersonToNotify                                 donor.Path
	RemoveReplacementAttorney                            donor.Path
	RemoveReplacementTrustCorporation                    donor.Path
	RemoveTrustCorporation                               donor.Path
	ResendCertificateProviderCode                        donor.Path
	ResendIndependentWitnessCode                         donor.Path
	ResendVoucherAccessCode                              donor.Path
	Restrictions                                         donor.Path
	SendUsYourEvidenceByPost                             donor.Path
	SignTheLpaOnBehalf                                   donor.Path
	SignYourLpa                                          donor.Path
	TaskList                                             donor.Path
	UnableToConfirmIdentity                              donor.Path
	UploadEvidence                                       donor.Path
	UploadEvidenceSSE                                    donor.Path
	UseExistingAddress                                   donor.Path
	ViewLPA                                              donor.Path
	WarningInterruption                                  donor.Path
	WeHaveContactedVoucher                               donor.Path
	WeHaveUpdatedYourDetails                             donor.Path
	WhatACertificateProviderDoes                         donor.Path
	WhatCountryDoYouLiveIn                               donor.Path
	WhatHappensNextRegisteringWithCourtOfProtection      donor.Path
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
	YourMobile                                           donor.Path
	YourName                                             donor.Path
	YourNonUKAddress                                     donor.Path
	YourPreferredLanguage                                donor.Path
}

var paths = appPaths{
	CertificateProvider: certificateProviderPaths{
		ConfirmDontWantToBeCertificateProviderLoggedOut: page.PathCertificateProviderConfirmDontWantToBeCertificateProviderLoggedOut,
		EnterReferenceNumber:                            page.PathCertificateProviderEnterReferenceNumber,
		EnterReferenceNumberOptOut:                      page.PathCertificateProviderEnterReferenceNumberOptOut,
		Login:                                           page.PathCertificateProviderLogin,
		LoginCallback:                                   page.PathCertificateProviderLoginCallback,
		YouHaveAlreadyProvidedACertificate:              page.PathCertificateProviderYouHaveAlreadyProvidedACertificate,
		YouHaveAlreadyProvidedACertificateLoggedIn:      page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn,
		YouHaveDecidedNotToBeCertificateProvider:        page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider,

		CertificateProvided:                    certificateprovider.PathCertificateProvided,
		ConfirmDontWantToBeCertificateProvider: certificateprovider.PathConfirmDontWantToBeCertificateProvider,
		ConfirmYourDetails:                     certificateprovider.PathConfirmYourDetails,
		EnterDateOfBirth:                       certificateprovider.PathEnterDateOfBirth,
		IdentityWithOneLogin:                   certificateprovider.PathIdentityWithOneLogin,
		IdentityWithOneLoginCallback:           certificateprovider.PathIdentityWithOneLoginCallback,
		ProveYourIdentity:                      certificateprovider.PathConfirmYourIdentity,
		OneLoginIdentityDetails:                certificateprovider.PathIdentityDetails,
		ProvideCertificate:                     certificateprovider.PathProvideCertificate,
		ReadTheLpa:                             certificateprovider.PathReadTheLpa,
		ReadTheDraftLpa:                        certificateprovider.PathReadTheDraftLpa,
		TaskList:                               certificateprovider.PathTaskList,
		WhatHappensNext:                        certificateprovider.PathWhatHappensNext,
		WhatIsYourHomeAddress:                  certificateprovider.PathWhatIsYourHomeAddress,
		WhoIsEligible:                          certificateprovider.PathWhoIsEligible,
		YourPreferredLanguage:                  certificateprovider.PathYourPreferredLanguage,
		YourRole:                               certificateprovider.PathYourRole,
	},

	Attorney: attorneyPaths{
		ConfirmDontWantToBeAttorneyLoggedOut: page.PathAttorneyConfirmDontWantToBeAttorneyLoggedOut,
		EnterReferenceNumber:                 page.PathAttorneyEnterReferenceNumber,
		EnterReferenceNumberOptOut:           page.PathAttorneyEnterReferenceNumberOptOut,
		Login:                                page.PathAttorneyLogin,
		LoginCallback:                        page.PathAttorneyLoginCallback,
		Start:                                page.PathAttorneyStart,
		YouHaveDecidedNotToBeAttorney:        page.PathAttorneyYouHaveDecidedNotToBeAttorney,

		CodeOfConduct:               attorney.PathCodeOfConduct,
		ConfirmDontWantToBeAttorney: attorney.PathConfirmDontWantToBeAttorney,
		ConfirmYourDetails:          attorney.PathConfirmYourDetails,
		PhoneNumber:                 attorney.PathPhoneNumber,
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

	Supporter: supporterPaths{
		EnterOrganisationName: page.PathSupporterEnterOrganisationName,
		EnterReferenceNumber:  page.PathSupporterEnterReferenceNumber,
		EnterYourName:         page.PathSupporterEnterYourName,
		Login:                 page.PathSupporterLogin,
		LoginCallback:         page.PathSupporterLoginCallback,
		OrganisationDeleted:   page.PathSupporterOrganisationDeleted,
		SigningInAdvice:       page.PathSupporterSigningInAdvice,
		Start:                 page.PathSupporterStart,
		InviteExpired:         page.PathSupporterInviteExpired,

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

	Voucher: voucherPaths{
		EnterReferenceNumber: page.PathVoucherEnterReferenceNumber,
		Login:                page.PathVoucherLogin,
		Start:                page.PathVoucherStart,

		TaskList:             voucher.PathTaskList,
		YourName:             voucher.PathYourName,
		IdentityWithOneLogin: voucher.PathIdentityWithOneLogin,
	},

	HealthCheck: healthCheckPaths{
		Service:    page.PathHealthCheckService,
		Dependency: page.PathHealthCheckDependency,
	},

	AccessibilityStatement:      page.PathAccessibilityStatement,
	AttorneyFixtures:            page.PathAttorneyFixtures,
	AddAnLPA:                    page.PathAddAnLPA,
	AuthRedirect:                page.PathAuthRedirect,
	CertificateProviderFixtures: page.PathCertificateProviderFixtures,
	CertificateProviderStart:    page.PathCertificateProviderStart,
	CookiesConsent:              page.PathCookiesConsent,
	Dashboard:                   page.PathDashboard,
	DashboardFixtures:           page.PathDashboardFixtures,
	EnterAccessCode:             page.PathEnterAccessCode,
	Fixtures:                    page.PathFixtures,
	Login:                       page.PathLogin,
	LoginCallback:               page.PathLoginCallback,
	LpaDeleted:                  page.PathLpaDeleted,
	LpaWithdrawn:                page.PathLpaWithdrawn,
	MakeOrAddAnLPA:              page.PathMakeOrAddAnLPA,
	Root:                        page.PathRoot,
	SignOut:                     page.PathSignOut,
	Start:                       page.PathStart,
	SupporterFixtures:           page.PathSupporterFixtures,
	VoucherFixtures:             page.PathVoucherFixtures,
	VoucherStart:                page.PathVoucherStart,

	AddingRestrictionsAndConditions:          page.PathAddingRestrictionsAndConditions,
	ContactTheOfficeOfThePublicGuardian:      page.PathContactTheOfficeOfThePublicGuardian,
	Glossary:                                 page.PathGlossary,
	HowDecisionsAreMadeWithMultipleAttorneys: page.PathHowDecisionsAreMadeWithMultipleAttorneys,
	HowToMakeAndRegisterYourLPA:              page.PathHowToMakeAndRegisterYourLPA,
	HowToSelectAttorneysForAnLPA:             page.PathHowToSelectAttorneysForAnLPA,
	ReplacementAttorneys:                     page.PathReplacementAttorneys,
	TheTwoTypesOfLPA:                         page.PathTheTwoTypesOfLPAPath,
	UnderstandingLifeSustainingTreatment:     page.PathUnderstandingLifeSustainingTreatment,
	UnderstandingMentalCapacity:              page.PathUnderstandingMentalCapacity,

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
	ChangeDonorEmail:                                     donor.PathChangeDonorEmail,
	ChangeDonorMobileNumber:                              donor.PathChangeDonorMobileNumber,
	ChangeIndependentWitnessMobileNumber:                 donor.PathChangeIndependentWitnessMobileNumber,
	CheckYouCanSign:                                      donor.PathCheckYouCanSign,
	CheckYourDetails:                                     donor.PathCheckYourDetails,
	CheckYourLpa:                                         donor.PathCheckYourLpa,
	ChooseAttorneys:                                      donor.PathChooseAttorneys,
	ChooseAttorneysAddress:                               donor.PathChooseAttorneysAddress,
	ChooseAttorneysGuidance:                              donor.PathChooseAttorneysGuidance,
	ChooseAttorneysSummary:                               donor.PathChooseAttorneysSummary,
	ChooseCertificateProvider:                            donor.PathChooseCertificateProvider,
	ChooseNewCertificateProvider:                         donor.PathChooseNewCertificateProvider,
	ChoosePeopleToNotify:                                 donor.PathEnterPersonToNotify,
	ChoosePeopleToNotifyAddress:                          donor.PathEnterPersonToNotifyAddress,
	ChoosePeopleToNotifySummary:                          donor.PathChoosePeopleToNotifySummary,
	ChooseReplacementAttorneys:                           donor.PathChooseReplacementAttorneys,
	ChooseReplacementAttorneysAddress:                    donor.PathChooseReplacementAttorneysAddress,
	ChooseReplacementAttorneysSummary:                    donor.PathChooseReplacementAttorneysSummary,
	ChooseReplacementTrustCorporation:                    donor.PathChooseReplacementTrustCorporation,
	ChooseSomeoneToVouchForYou:                           donor.PathChooseSomeoneToVouchForYou,
	ChooseTrustCorporation:                               donor.PathChooseTrustCorporation,
	ChooseYourCertificateProvider:                        donor.PathChooseYourCertificateProvider,
	ConfirmPersonAllowedToVouch:                          donor.PathConfirmPersonAllowedToVouch,
	ConfirmYourCertificateProviderIsNotRelated:           donor.PathConfirmYourCertificateProviderIsNotRelated,
	DeleteThisLpa:                                        donor.PathDeleteThisLpa,
	DoYouLiveInTheUK:                                     donor.PathDoYouLiveInTheUK,
	DoYouWantReplacementAttorneys:                        donor.PathDoYouWantReplacementAttorneys,
	DoYouWantToNotifyPeople:                              donor.PathDoYouWantToNotifyPeople,
	EnterAttorney:                                        donor.PathEnterAttorney,
	EnterCorrespondentAddress:                            donor.PathEnterCorrespondentAddress,
	EnterCorrespondentDetails:                            donor.PathEnterCorrespondentDetails,
	EnterReplacementAttorney:                             donor.PathEnterReplacementAttorney,
	EnterReplacementTrustCorporation:                     donor.PathEnterReplacementTrustCorporation,
	EnterReplacementTrustCorporationAddress:              donor.PathEnterReplacementTrustCorporationAddress,
	EnterTrustCorporation:                                donor.PathEnterTrustCorporation,
	EnterTrustCorporationAddress:                         donor.PathEnterTrustCorporationAddress,
	EnterVoucher:                                         donor.PathEnterVoucher,
	EvidenceRequired:                                     donor.PathEvidenceRequired,
	EvidenceSuccessfullyUploaded:                         donor.PathEvidenceSuccessfullyUploaded,
	GettingHelpSigning:                                   donor.PathGettingHelpSigning,
	HowDoYouKnowYourCertificateProvider:                  donor.PathHowDoYouKnowYourCertificateProvider,
	HowLongHaveYouKnownCertificateProvider:               donor.PathHowLongHaveYouKnownCertificateProvider,
	HowShouldAttorneysMakeDecisions:                      donor.PathHowShouldAttorneysMakeDecisions,
	HowShouldReplacementAttorneysMakeDecisions:           donor.PathHowShouldReplacementAttorneysMakeDecisions,
	HowShouldReplacementAttorneysStepIn:                  donor.PathHowShouldReplacementAttorneysStepIn,
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
	OneLoginIdentityDetails:                              donor.PathIdentityDetails,
	PayFee:                                               donor.PathPayFee,
	PaymentConfirmation:                                  donor.PathPaymentConfirmation,
	PreviousApplicationNumber:                            donor.PathPreviousApplicationNumber,
	PreviousFee:                                          donor.PathPreviousFee,
	Progress:                                             donor.PathProgress,
	ProveYourIdentity:                                    donor.PathConfirmYourIdentity,
	ReadYourLpa:                                          donor.PathReadYourLpa,
	ReceivingUpdatesAboutYourLpa:                         donor.PathReceivingUpdatesAboutYourLpa,
	RegisterWithCourtOfProtection:                        donor.PathRegisterWithCourtOfProtection,
	RemoveAttorney:                                       donor.PathRemoveAttorney,
	RemoveCertificateProvider:                            donor.PathRemoveCertificateProvider,
	RemoveCorrespondent:                                  donor.PathRemoveCorrespondent,
	RemovePersonToNotify:                                 donor.PathRemovePersonToNotify,
	RemoveReplacementAttorney:                            donor.PathRemoveReplacementAttorney,
	RemoveReplacementTrustCorporation:                    donor.PathRemoveReplacementTrustCorporation,
	RemoveTrustCorporation:                               donor.PathRemoveTrustCorporation,
	ResendCertificateProviderCode:                        donor.PathResendCertificateProviderCode,
	ResendIndependentWitnessCode:                         donor.PathResendIndependentWitnessCode,
	ResendVoucherAccessCode:                              donor.PathResendVoucherAccessCode,
	Restrictions:                                         donor.PathRestrictions,
	SendUsYourEvidenceByPost:                             donor.PathSendUsYourEvidenceByPost,
	SignTheLpaOnBehalf:                                   donor.PathSignTheLpaOnBehalf,
	SignYourLpa:                                          donor.PathSignYourLpa,
	TaskList:                                             donor.PathTaskList,
	UnableToConfirmIdentity:                              donor.PathUnableToConfirmIdentity,
	UploadEvidence:                                       donor.PathUploadEvidence,
	UploadEvidenceSSE:                                    donor.PathUploadEvidenceSSE,
	UseExistingAddress:                                   donor.PathUseExistingAddress,
	ViewLPA:                                              donor.PathViewLPA,
	WarningInterruption:                                  donor.PathWarningInterruption,
	WeHaveContactedVoucher:                               donor.PathWeHaveContactedVoucher,
	WeHaveUpdatedYourDetails:                             donor.PathWeHaveUpdatedYourDetails,
	WhatACertificateProviderDoes:                         donor.PathWhatACertificateProviderDoes,
	WhatCountryDoYouLiveIn:                               donor.PathWhatCountryDoYouLiveIn,
	WhatHappensNextRegisteringWithCourtOfProtection:      donor.PathWhatHappensNextRegisteringWithCourtOfProtection,
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
	YourMobile:                                           donor.PathYourMobile,
	YourName:                                             donor.PathYourName,
	YourNonUKAddress:                                     donor.PathYourNonUKAddress,
	YourPreferredLanguage:                                donor.PathYourPreferredLanguage,
}
