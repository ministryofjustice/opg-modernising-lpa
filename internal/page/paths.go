package page

type Path string

func (p Path) String() string {
	return string(p)
}

func (p Path) Format() string {
	return string(p)
}

type LpaPath string

func (p LpaPath) String() string {
	return string(p)
}

func (p LpaPath) Format(id string) string {
	return "/lpa/" + id + string(p)
}

type AttorneyPath string

func (p AttorneyPath) String() string {
	return string(p)
}

func (p AttorneyPath) Format(id string) string {
	return "/attorney/" + id + string(p)
}

type CertificateProviderPath string

func (p CertificateProviderPath) String() string {
	return string(p)
}

func (p CertificateProviderPath) Format(id string) string {
	return "/certificate-provider/" + id + string(p)
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
}

type CertificateProviderPaths struct {
	Login                Path
	LoginCallback        Path
	EnterReferenceNumber Path

	CertificateProvided                  CertificateProviderPath
	ConfirmYourDetails                   CertificateProviderPath
	EnterDateOfBirth                     CertificateProviderPath
	IdentityWithBiometricResidencePermit CertificateProviderPath
	IdentityWithDrivingLicencePaper      CertificateProviderPath
	IdentityWithDrivingLicencePhotocard  CertificateProviderPath
	IdentityWithOneLogin                 CertificateProviderPath
	IdentityWithOneLoginCallback         CertificateProviderPath
	IdentityWithOnlineBankAccount        CertificateProviderPath
	IdentityWithPassport                 CertificateProviderPath
	IdentityWithYoti                     CertificateProviderPath
	IdentityWithYotiCallback             CertificateProviderPath
	ProvideCertificate                   CertificateProviderPath
	ReadTheLpa                           CertificateProviderPath
	SelectYourIdentityOptions            CertificateProviderPath
	SelectYourIdentityOptions1           CertificateProviderPath
	SelectYourIdentityOptions2           CertificateProviderPath
	TaskList                             CertificateProviderPath
	WhatHappensNext                      CertificateProviderPath
	WhatYoullNeedToConfirmYourIdentity   CertificateProviderPath
	WhoIsEligible                        CertificateProviderPath
	YourChosenIdentityOptions            CertificateProviderPath
	YourRole                             CertificateProviderPath
}

type HealthCheckPaths struct {
	Service    Path
	Dependency Path
}

type AppPaths struct {
	Attorney            AttorneyPaths
	CertificateProvider CertificateProviderPaths
	HealthCheck         HealthCheckPaths

	AuthRedirect                       Path
	Login                              Path
	LoginCallback                      Path
	Root                               Path
	SignOut                            Path
	Fixtures                           Path
	AttorneyFixtures                   Path
	CertificateProviderFixtures        Path
	DashboardFixtures                  Path
	YourLegalRightsAndResponsibilities Path
	CertificateProviderStart           Path
	Start                              Path
	Dashboard                          Path
	YotiRedirect                       Path
	CookiesConsent                     Path

	AboutPayment                                         LpaPath
	ApplicationReason                                    LpaPath
	AreYouApplyingForFeeDiscountOrExemption              LpaPath
	CertificateProviderAddress                           LpaPath
	CertificateProviderDetails                           LpaPath
	CertificateProviderOptOut                            LpaPath
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
	DoYouWantReplacementAttorneys                        LpaPath
	DoYouWantToNotifyPeople                              LpaPath
	EnterReplacementTrustCorporation                     LpaPath
	EnterReplacementTrustCorporationAddress              LpaPath
	EnterTrustCorporation                                LpaPath
	EnterTrustCorporationAddress                         LpaPath
	EvidenceRequired                                     LpaPath
	FeeDenied                                            LpaPath
	GettingHelpSigning                                   LpaPath
	HowDoYouKnowYourCertificateProvider                  LpaPath
	HowLongHaveYouKnownCertificateProvider               LpaPath
	HowShouldAttorneysMakeDecisions                      LpaPath
	HowShouldReplacementAttorneysMakeDecisions           LpaPath
	HowShouldReplacementAttorneysStepIn                  LpaPath
	HowToConfirmYourIdentityAndSign                      LpaPath
	HowToEmailOrPostEvidence                             LpaPath
	HowToSendEvidence                                    LpaPath
	HowWouldCertificateProviderPreferToCarryOutTheirRole LpaPath
	HowWouldYouLikeToSendEvidence                        LpaPath
	IdentityConfirmed                                    LpaPath
	IdentityWithBiometricResidencePermit                 LpaPath
	IdentityWithDrivingLicencePaper                      LpaPath
	IdentityWithDrivingLicencePhotocard                  LpaPath
	IdentityWithOneLogin                                 LpaPath
	IdentityWithOneLoginCallback                         LpaPath
	IdentityWithOnlineBankAccount                        LpaPath
	IdentityWithPassport                                 LpaPath
	IdentityWithYoti                                     LpaPath
	IdentityWithYotiCallback                             LpaPath
	LifeSustainingTreatment                              LpaPath
	LpaDetailsSaved                                      LpaPath
	LpaType                                              LpaPath
	LpaYourLegalRightsAndResponsibilities                LpaPath
	NeedHelpSigningConfirmation                          LpaPath
	PaymentConfirmation                                  LpaPath
	PreviousApplicationNumber                            LpaPath
	Progress                                             LpaPath
	ReadYourLpa                                          LpaPath
	RemoveAttorney                                       LpaPath
	RemovePersonToNotify                                 LpaPath
	RemoveReplacementAttorney                            LpaPath
	ResendCertificateProviderCode                        LpaPath
	ResendIndependentWitnessCode                         LpaPath
	Restrictions                                         LpaPath
	SelectYourIdentityOptions                            LpaPath
	SelectYourIdentityOptions1                           LpaPath
	SelectYourIdentityOptions2                           LpaPath
	SignTheLpaOnBehalf                                   LpaPath
	SignYourLpa                                          LpaPath
	TaskList                                             LpaPath
	UploadEvidence                                       LpaPath
	UseExistingAddress                                   LpaPath
	WhatACertificateProviderDoes                         LpaPath
	WhatHappensAfterNoFee                                LpaPath
	WhatYoullNeedToConfirmYourIdentity                   LpaPath
	WhenCanTheLpaBeUsed                                  LpaPath
	WhichFeeTypeAreYouApplyingFor                        LpaPath
	WitnessingAsCertificateProvider                      LpaPath
	WitnessingAsIndependentWitness                       LpaPath
	WitnessingYourSignature                              LpaPath
	YouHaveSubmittedYourLpa                              LpaPath
	YourAddress                                          LpaPath
	YourAuthorisedSignatory                              LpaPath
	YourChosenIdentityOptions                            LpaPath
	YourDetails                                          LpaPath
	YourIndependentWitness                               LpaPath
	YourIndependentWitnessAddress                        LpaPath
	YourIndependentWitnessMobile                         LpaPath
}

var Paths = AppPaths{
	CertificateProvider: CertificateProviderPaths{
		Login:                "/certificate-provider-login",
		LoginCallback:        "/certificate-provider-login-callback",
		EnterReferenceNumber: "/certificate-provider-enter-reference-number",
		WhoIsEligible:        "/certificate-provider-who-is-eligible",

		CertificateProvided:                  "/certificate-provided",
		ConfirmYourDetails:                   "/confirm-your-details",
		EnterDateOfBirth:                     "/enter-date-of-birth",
		IdentityWithBiometricResidencePermit: "/id/brp",
		IdentityWithDrivingLicencePaper:      "/id/dlpaper",
		IdentityWithDrivingLicencePhotocard:  "/id/dlphoto",
		IdentityWithOneLogin:                 "/identity-with-one-login",
		IdentityWithOneLoginCallback:         "/identity-with-one-login-callback",
		IdentityWithOnlineBankAccount:        "/id/bank",
		IdentityWithPassport:                 "/id/passport",
		IdentityWithYoti:                     "/identity-with-yoti",
		IdentityWithYotiCallback:             "/identity-with-yoti-callback",
		ProvideCertificate:                   "/provide-certificate",
		ReadTheLpa:                           "/read-the-lpa",
		SelectYourIdentityOptions1:           "/select-identity-document",
		SelectYourIdentityOptions2:           "/select-identity-document-2",
		SelectYourIdentityOptions:            "/select-your-identity-options",
		TaskList:                             "/task-list",
		WhatHappensNext:                      "/what-happens-next",
		WhatYoullNeedToConfirmYourIdentity:   "/what-youll-need-to-confirm-your-identity",
		YourChosenIdentityOptions:            "/your-chosen-identity-options",
		YourRole:                             "/your-role",
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
	},

	HealthCheck: HealthCheckPaths{
		Service:    "/health-check/service",
		Dependency: "/health-check/dependency",
	},

	AboutPayment:                                         "/about-payment",
	ApplicationReason:                                    "/application-reason",
	AreYouApplyingForFeeDiscountOrExemption:              "/are-you-applying-for-fee-discount-or-exemption",
	AttorneyFixtures:                                     "/fixtures/attorney",
	AuthRedirect:                                         "/auth/redirect",
	CertificateProviderAddress:                           "/certificate-provider-address",
	CertificateProviderDetails:                           "/certificate-provider-details",
	CertificateProviderFixtures:                          "/fixtures/certificate-provider",
	CertificateProviderOptOut:                            "/certificate-provider-opt-out",
	CertificateProviderStart:                             "/certificate-provider-start",
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
	CookiesConsent:                                       "/cookies-consent",
	Dashboard:                                            "/dashboard",
	DashboardFixtures:                                    "/fixtures/dashboard",
	DoYouWantReplacementAttorneys:                        "/do-you-want-replacement-attorneys",
	DoYouWantToNotifyPeople:                              "/do-you-want-to-notify-people",
	EnterReplacementTrustCorporation:                     "/enter-replacement-trust-corporation",
	EnterReplacementTrustCorporationAddress:              "/enter-replacement-trust-corporation-address",
	EnterTrustCorporation:                                "/enter-trust-corporation",
	EnterTrustCorporationAddress:                         "/enter-trust-corporation-address",
	EvidenceRequired:                                     "/evidence-required",
	FeeDenied:                                            "/fee-denied",
	Fixtures:                                             "/fixtures",
	GettingHelpSigning:                                   "/getting-help-signing",
	HowDoYouKnowYourCertificateProvider:                  "/how-do-you-know-your-certificate-provider",
	HowLongHaveYouKnownCertificateProvider:               "/how-long-have-you-known-certificate-provider",
	HowShouldAttorneysMakeDecisions:                      "/how-should-attorneys-make-decisions",
	HowShouldReplacementAttorneysMakeDecisions:           "/how-should-replacement-attorneys-make-decisions",
	HowShouldReplacementAttorneysStepIn:                  "/how-should-replacement-attorneys-step-in",
	HowToConfirmYourIdentityAndSign:                      "/how-to-confirm-your-identity-and-sign",
	HowToEmailOrPostEvidence:                             "/how-to-email-or-post-evidence",
	HowToSendEvidence:                                    "/how-to-send-evidence",
	HowWouldCertificateProviderPreferToCarryOutTheirRole: "/how-would-certificate-provider-prefer-to-carry-out-their-role",
	HowWouldYouLikeToSendEvidence:                        "/how-would-you-like-to-send-evidence",
	IdentityConfirmed:                                    "/identity-confirmed",
	IdentityWithBiometricResidencePermit:                 "/id/biometric-residence-permit",
	IdentityWithDrivingLicencePaper:                      "/id/driving-licence-paper",
	IdentityWithDrivingLicencePhotocard:                  "/id/driving-licence-photocard",
	IdentityWithOneLogin:                                 "/id/one-login",
	IdentityWithOneLoginCallback:                         "/id/one-login/callback",
	IdentityWithOnlineBankAccount:                        "/id/online-bank-account",
	IdentityWithPassport:                                 "/id/passport",
	IdentityWithYoti:                                     "/id/yoti",
	IdentityWithYotiCallback:                             "/id/yoti/callback",
	LifeSustainingTreatment:                              "/life-sustaining-treatment",
	Login:                                                "/login",
	LoginCallback:                                        "/login-callback",
	LpaDetailsSaved:                                      "/lpa-details-saved",
	LpaType:                                              "/lpa-type",
	LpaYourLegalRightsAndResponsibilities:                "/your-legal-rights-and-responsibilities",
	NeedHelpSigningConfirmation:                          "/need-help-signing-confirmation",
	PaymentConfirmation:                                  "/payment-confirmation",
	PreviousApplicationNumber:                            "/previous-application-number",
	Progress:                                             "/progress",
	ReadYourLpa:                                          "/read-your-lpa",
	RemoveAttorney:                                       "/remove-attorney",
	RemovePersonToNotify:                                 "/remove-person-to-notify",
	RemoveReplacementAttorney:                            "/remove-replacement-attorney",
	ResendCertificateProviderCode:                        "/resend-certificate-provider-code",
	ResendIndependentWitnessCode:                         "/resend-independent-witness-code",
	Restrictions:                                         "/restrictions",
	Root:                                                 "/",
	SelectYourIdentityOptions1:                           "/select-identity-document",
	SelectYourIdentityOptions2:                           "/select-identity-document-2",
	SelectYourIdentityOptions:                            "/select-your-identity-options",
	SignOut:                                              "/sign-out",
	SignTheLpaOnBehalf:                                   "/sign-the-lpa-on-behalf",
	SignYourLpa:                                          "/sign-your-lpa",
	Start:                                                "/start",
	TaskList:                                             "/task-list",
	UploadEvidence:                                       "/upload-evidence",
	UseExistingAddress:                                   "/use-existing-address",
	WhatACertificateProviderDoes:                         "/what-a-certificate-provider-does",
	WhatHappensAfterNoFee:                                "/what-happens-after-no-fee",
	WhatYoullNeedToConfirmYourIdentity:                   "/what-youll-need-to-confirm-your-identity",
	WhenCanTheLpaBeUsed:                                  "/when-can-the-lpa-be-used",
	WhichFeeTypeAreYouApplyingFor:                        "/which-fee-type-are-you-applying-for",
	WitnessingAsCertificateProvider:                      "/witnessing-as-certificate-provider",
	WitnessingAsIndependentWitness:                       "/witnessing-as-independent-witness",
	WitnessingYourSignature:                              "/witnessing-your-signature",
	YotiRedirect:                                         "/yoti/redirect",
	YouHaveSubmittedYourLpa:                              "/you-have-submitted-your-lpa",
	YourAddress:                                          "/your-address",
	YourAuthorisedSignatory:                              "/your-authorised-signatory",
	YourChosenIdentityOptions:                            "/your-chosen-identity-options",
	YourDetails:                                          "/your-details",
	YourIndependentWitness:                               "/your-independent-witness",
	YourIndependentWitnessAddress:                        "/your-independent-witness-address",
	YourIndependentWitnessMobile:                         "/your-independent-witness-mobile",
	YourLegalRightsAndResponsibilities:                   "/your-legal-rights-and-responsibilities",
}
