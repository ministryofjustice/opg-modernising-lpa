package page

import (
	"strings"

	"golang.org/x/exp/slices"
)

type AttorneyPaths struct {
	Start                     string
	Login                     string
	LoginCallback             string
	EnterReferenceNumber      string
	CodeOfConduct             string
	TaskList                  string
	CheckYourName             string
	DateOfBirth               string
	MobileNumber              string
	YourAddress               string
	ReadTheLpa                string
	RightsAndResponsibilities string
	WhatHappensWhenYouSign    string
	Sign                      string
	WhatHappensNext           string
}

type AppPaths struct {
	AboutPayment                                               string
	AreYouHappyIfOneAttorneyCantActNoneCan                     string
	AreYouHappyIfOneReplacementAttorneyCantActNoneCan          string
	AreYouHappyIfRemainingAttorneysCanContinueToAct            string
	AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct string
	Attorney                                                   AttorneyPaths
	AuthRedirect                                               string
	CertificateProvided                                        string
	CertificateProviderAddress                                 string
	CertificateProviderCheckYourName                           string
	CertificateProviderDetails                                 string
	CertificateProviderEnterDateOfBirth                        string
	CertificateProviderEnterMobileNumber                       string
	CertificateProviderEnterReferenceNumber                    string
	CertificateProviderIdentityWithBiometricResidencePermit    string
	CertificateProviderIdentityWithDrivingLicencePaper         string
	CertificateProviderIdentityWithDrivingLicencePhotocard     string
	CertificateProviderIdentityWithOneLogin                    string
	CertificateProviderIdentityWithOneLoginCallback            string
	CertificateProviderIdentityWithOnlineBankAccount           string
	CertificateProviderIdentityWithPassport                    string
	CertificateProviderIdentityWithYoti                        string
	CertificateProviderIdentityWithYotiCallback                string
	CertificateProviderLogin                                   string
	CertificateProviderLoginCallback                           string
	CertificateProviderOptOut                                  string
	CertificateProviderReadTheLpa                              string
	CertificateProviderSelectYourIdentityOptions               string
	CertificateProviderSelectYourIdentityOptions1              string
	CertificateProviderSelectYourIdentityOptions2              string
	CertificateProviderStart                                   string
	CertificateProviderWhatHappensNext                         string
	CertificateProviderWhatYoullNeedToConfirmYourIdentity      string
	CertificateProviderWhoIsEligible                           string
	CertificateProviderYourChosenIdentityOptions               string
	CheckYourLpa                                               string
	ChooseAttorneys                                            string
	ChooseAttorneysAddress                                     string
	ChooseAttorneysSummary                                     string
	ChoosePeopleToNotify                                       string
	ChoosePeopleToNotifyAddress                                string
	ChoosePeopleToNotifySummary                                string
	ChooseReplacementAttorneys                                 string
	ChooseReplacementAttorneysAddress                          string
	ChooseReplacementAttorneysSummary                          string
	CookiesConsent                                             string
	Dashboard                                                  string
	DoYouWantReplacementAttorneys                              string
	DoYouWantToNotifyPeople                                    string
	Fixtures                                                   string
	HealthCheck                                                string
	HowDoYouKnowYourCertificateProvider                        string
	HowLongHaveYouKnownCertificateProvider                     string
	HowShouldAttorneysMakeDecisions                            string
	HowShouldReplacementAttorneysMakeDecisions                 string
	HowShouldReplacementAttorneysStepIn                        string
	HowToConfirmYourIdentityAndSign                            string
	HowWouldCertificateProviderPreferToCarryOutTheirRole       string
	IdentityConfirmed                                          string
	IdentityWithBiometricResidencePermit                       string
	IdentityWithDrivingLicencePaper                            string
	IdentityWithDrivingLicencePhotocard                        string
	IdentityWithOneLogin                                       string
	IdentityWithOneLoginCallback                               string
	IdentityWithOnlineBankAccount                              string
	IdentityWithPassport                                       string
	IdentityWithYoti                                           string
	IdentityWithYotiCallback                                   string
	LifeSustainingTreatment                                    string
	Login                                                      string
	LoginCallback                                              string
	LpaType                                                    string
	PaymentConfirmation                                        string
	Progress                                                   string
	ProvideCertificate                                         string
	ReadYourLpa                                                string
	RemoveAttorney                                             string
	RemovePersonToNotify                                       string
	RemoveReplacementAttorney                                  string
	ResendWitnessCode                                          string
	Restrictions                                               string
	Root                                                       string
	SelectYourIdentityOptions                                  string
	SelectYourIdentityOptions1                                 string
	SelectYourIdentityOptions2                                 string
	SignYourLpa                                                string
	Start                                                      string
	TaskList                                                   string
	TestingStart                                               string
	UseExistingAddress                                         string
	WhatYoullNeedToConfirmYourIdentity                         string
	WhenCanTheLpaBeUsed                                        string
	WhoDoYouWantToBeCertificateProviderGuidance                string
	WhoIsTheLpaFor                                             string
	WitnessingAsCertificateProvider                            string
	WitnessingYourSignature                                    string
	YotiRedirect                                               string
	YouHaveSubmittedYourLpa                                    string
	YourAddress                                                string
	YourChosenIdentityOptions                                  string
	YourDetails                                                string
	YourLegalRightsAndResponsibilities                         string
}

var Paths = AppPaths{
	AboutPayment:                                               "/about-payment",
	AreYouHappyIfOneAttorneyCantActNoneCan:                     "/are-you-happy-if-one-attorney-cant-act-none-can",
	AreYouHappyIfOneReplacementAttorneyCantActNoneCan:          "/are-you-happy-if-one-replacement-attorney-cant-act-none-can",
	AreYouHappyIfRemainingAttorneysCanContinueToAct:            "/are-you-happy-if-remaining-attorneys-can-continue-to-act",
	AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct: "/are-you-happy-if-remaining-replacement-attorneys-can-continue-to-act",
	Attorney: AttorneyPaths{
		Start:                     "/attorney-start",
		Login:                     "/attorney-login",
		LoginCallback:             "/attorney-login-callback",
		TaskList:                  "/attorney-task-list",
		EnterReferenceNumber:      "/attorney-enter-reference-number",
		CodeOfConduct:             "/attorney-code-of-conduct",
		CheckYourName:             "/attorney-check-your-name",
		DateOfBirth:               "/attorney-date-of-birth",
		MobileNumber:              "/attorney-mobile-number",
		YourAddress:               "/attorney-your-address",
		ReadTheLpa:                "/attorney-read-the-lpa",
		RightsAndResponsibilities: "/attorney-legal-rights-and-responsibilities",
		WhatHappensWhenYouSign:    "/attorney-what-happens-when-you-sign-the-lpa",
		Sign:                      "/attorney-sign",
		WhatHappensNext:           "/attorney-what-happens-next",
	},
	AuthRedirect:                                            "/auth/redirect",
	CertificateProvided:                                     "/certificate-provided",
	CertificateProviderAddress:                              "/certificate-provider-address",
	CertificateProviderCheckYourName:                        "/certificate-provider-check-your-name",
	CertificateProviderDetails:                              "/certificate-provider-details",
	CertificateProviderEnterDateOfBirth:                     "/certificate-provider-enter-date-of-birth",
	CertificateProviderEnterMobileNumber:                    "/certificate-provider-enter-mobile-number",
	CertificateProviderEnterReferenceNumber:                 "/certificate-provider-enter-reference-number",
	CertificateProviderIdentityWithBiometricResidencePermit: "/certificate-provider/id/brp",
	CertificateProviderIdentityWithDrivingLicencePaper:      "/certificate-provider/id/dlpaper",
	CertificateProviderIdentityWithDrivingLicencePhotocard:  "/certificate-provider/id/dlphoto",
	CertificateProviderIdentityWithOneLogin:                 "/certificate-provider-identity-with-one-login",
	CertificateProviderIdentityWithOneLoginCallback:         "/certificate-provider-identity-with-one-login-callback",
	CertificateProviderIdentityWithOnlineBankAccount:        "/certificate-provider/id/bank",
	CertificateProviderIdentityWithPassport:                 "/certificate-provider/id/passport",
	CertificateProviderIdentityWithYoti:                     "/certificate-provider-identity-with-yoti",
	CertificateProviderIdentityWithYotiCallback:             "/certificate-provider-identity-with-yoti-callback",
	CertificateProviderLogin:                                "/certificate-provider-login",
	CertificateProviderLoginCallback:                        "/certificate-provider-login-callback",
	CertificateProviderOptOut:                               "/certificate-provider-opt-out",
	CertificateProviderReadTheLpa:                           "/certificate-provider-read-the-lpa",
	CertificateProviderSelectYourIdentityOptions1:           "/certificate-provider-select-identity-document",
	CertificateProviderSelectYourIdentityOptions2:           "/certificate-provider-select-identity-document-2",
	CertificateProviderSelectYourIdentityOptions:            "/certificate-provider-select-your-identity-options",
	CertificateProviderStart:                                "/certificate-provider-start",
	CertificateProviderWhatHappensNext:                      "/certificate-provider-what-happens-next",
	CertificateProviderWhatYoullNeedToConfirmYourIdentity:   "/certificate-provider-what-youll-need-to-confirm-your-identity",
	CertificateProviderWhoIsEligible:                        "/certificate-provider-who-is-eligible",
	CertificateProviderYourChosenIdentityOptions:            "/certificate-provider-your-chosen-identity-options",
	CheckYourLpa:                                         "/check-your-lpa",
	ChooseAttorneys:                                      "/choose-attorneys",
	ChooseAttorneysAddress:                               "/choose-attorneys-address",
	ChooseAttorneysSummary:                               "/choose-attorneys-summary",
	ChoosePeopleToNotify:                                 "/choose-people-to-notify",
	ChoosePeopleToNotifyAddress:                          "/choose-people-to-notify-address",
	ChoosePeopleToNotifySummary:                          "/choose-people-to-notify-summary",
	ChooseReplacementAttorneys:                           "/choose-replacement-attorneys",
	ChooseReplacementAttorneysAddress:                    "/choose-replacement-attorneys-address",
	ChooseReplacementAttorneysSummary:                    "/choose-replacement-attorneys-summary",
	CookiesConsent:                                       "/cookies-consent",
	Dashboard:                                            "/dashboard",
	DoYouWantReplacementAttorneys:                        "/do-you-want-replacement-attorneys",
	DoYouWantToNotifyPeople:                              "/do-you-want-to-notify-people",
	Fixtures:                                             "/fixtures",
	HealthCheck:                                          "/health-check",
	HowDoYouKnowYourCertificateProvider:                  "/how-do-you-know-your-certificate-provider",
	HowLongHaveYouKnownCertificateProvider:               "/how-long-have-you-known-certificate-provider",
	HowShouldAttorneysMakeDecisions:                      "/how-should-attorneys-make-decisions",
	HowShouldReplacementAttorneysMakeDecisions:           "/how-should-replacement-attorneys-make-decisions",
	HowShouldReplacementAttorneysStepIn:                  "/how-should-replacement-attorneys-step-in",
	HowToConfirmYourIdentityAndSign:                      "/how-to-confirm-your-identity-and-sign",
	HowWouldCertificateProviderPreferToCarryOutTheirRole: "/how-would-certificate-provider-prefer-to-carry-out-their-role",
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
	LpaType:                                              "/lpa-type",
	PaymentConfirmation:                                  "/payment-confirmation",
	Progress:                                             "/progress",
	ProvideCertificate:                                   "/provide-certificate",
	ReadYourLpa:                                          "/read-your-lpa",
	RemoveAttorney:                                       "/remove-attorney",
	RemovePersonToNotify:                                 "/remove-person-to-notify",
	RemoveReplacementAttorney:                            "/remove-replacement-attorney",
	ResendWitnessCode:                                    "/resend-witness-code",
	Restrictions:                                         "/restrictions",
	Root:                                                 "/",
	SelectYourIdentityOptions1:                           "/select-identity-document",
	SelectYourIdentityOptions2:                           "/select-identity-document-2",
	SelectYourIdentityOptions:                            "/select-your-identity-options",
	SignYourLpa:                                          "/sign-your-lpa",
	Start:                                                "/start",
	TaskList:                                             "/task-list",
	TestingStart:                                         "/testing-start",
	UseExistingAddress:                                   "/use-existing-address",
	WhatYoullNeedToConfirmYourIdentity:                   "/what-youll-need-to-confirm-your-identity",
	WhenCanTheLpaBeUsed:                                  "/when-can-the-lpa-be-used",
	WhoDoYouWantToBeCertificateProviderGuidance:          "/who-do-you-want-to-be-certificate-provider-guidance",
	WhoIsTheLpaFor:                                       "/who-is-the-lpa-for",
	WitnessingAsCertificateProvider:                      "/witnessing-as-certificate-provider",
	WitnessingYourSignature:                              "/witnessing-your-signature",
	YotiRedirect:                                         "/yoti/redirect",
	YouHaveSubmittedYourLpa:                              "/you-have-submitted-your-lpa",
	YourAddress:                                          "/your-address",
	YourChosenIdentityOptions:                            "/your-chosen-identity-options",
	YourDetails:                                          "/your-details",
	YourLegalRightsAndResponsibilities:                   "/your-legal-rights-and-responsibilities",
}

func IsLpaPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return !slices.Contains([]string{
		Paths.YotiRedirect,
		Paths.Attorney.YourAddress,
		Paths.Attorney.CheckYourName,
		Paths.Attorney.CodeOfConduct,
		Paths.Attorney.DateOfBirth,
		Paths.Attorney.EnterReferenceNumber,
		Paths.Attorney.Login,
		Paths.Attorney.LoginCallback,
		Paths.Attorney.MobileNumber,
		Paths.Attorney.ReadTheLpa,
		Paths.Attorney.RightsAndResponsibilities,
		Paths.Attorney.Sign,
		Paths.Attorney.Start,
		Paths.Attorney.TaskList,
		Paths.Attorney.WhatHappensWhenYouSign,
		Paths.Attorney.RightsAndResponsibilities,
		Paths.Attorney.WhatHappensNext,
		Paths.AuthRedirect,
		Paths.CertificateProvided,
		Paths.CertificateProviderCheckYourName,
		Paths.CertificateProviderEnterReferenceNumber,
		Paths.CertificateProviderIdentityWithBiometricResidencePermit,
		Paths.CertificateProviderIdentityWithDrivingLicencePaper,
		Paths.CertificateProviderIdentityWithDrivingLicencePhotocard,
		Paths.CertificateProviderIdentityWithOneLogin,
		Paths.CertificateProviderIdentityWithOneLoginCallback,
		Paths.CertificateProviderIdentityWithOnlineBankAccount,
		Paths.CertificateProviderIdentityWithPassport,
		Paths.CertificateProviderIdentityWithYoti,
		Paths.CertificateProviderIdentityWithYotiCallback,
		Paths.CertificateProviderLogin,
		Paths.CertificateProviderLoginCallback,
		Paths.CertificateProviderReadTheLpa,
		Paths.CertificateProviderSelectYourIdentityOptions,
		Paths.CertificateProviderSelectYourIdentityOptions1,
		Paths.CertificateProviderSelectYourIdentityOptions2,
		Paths.CertificateProviderStart,
		Paths.CertificateProviderWhatHappensNext,
		Paths.CertificateProviderWhatYoullNeedToConfirmYourIdentity,
		Paths.CertificateProviderYourChosenIdentityOptions,
		Paths.CertificateProviderEnterDateOfBirth,
		Paths.CertificateProviderEnterMobileNumber,
		Paths.CertificateProviderWhoIsEligible,
		Paths.Dashboard,
		Paths.Login,
		Paths.LoginCallback,
		Paths.ProvideCertificate,
		Paths.Start,
	}, path)
}
