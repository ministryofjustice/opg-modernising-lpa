package page

import (
	"strings"

	"golang.org/x/exp/slices"
)

type AttorneyPaths struct {
	CheckYourName             string
	CodeOfConduct             string
	EnterReferenceNumber      string
	Login                     string
	LoginCallback             string
	MobileNumber              string
	ReadTheLpa                string
	RightsAndResponsibilities string
	Sign                      string
	Start                     string
	TaskList                  string
	WhatHappensNext           string
	WhatHappensWhenYouSign    string
}

type CertificateProviderPaths struct {
	CertificateProvided                  string
	CheckYourName                        string
	EnterDateOfBirth                     string
	EnterMobileNumber                    string
	EnterReferenceNumber                 string
	IdentityWithBiometricResidencePermit string
	IdentityWithDrivingLicencePaper      string
	IdentityWithDrivingLicencePhotocard  string
	IdentityWithOneLogin                 string
	IdentityWithOneLoginCallback         string
	IdentityWithOnlineBankAccount        string
	IdentityWithPassport                 string
	IdentityWithYoti                     string
	IdentityWithYotiCallback             string
	Login                                string
	LoginCallback                        string
	ProvideCertificate                   string
	ReadTheLpa                           string
	SelectYourIdentityOptions            string
	SelectYourIdentityOptions1           string
	SelectYourIdentityOptions2           string
	WhatHappensNext                      string
	WhatYoullNeedToConfirmYourIdentity   string
	WhoIsEligible                        string
	YourChosenIdentityOptions            string
}
type HealthCheckPaths struct {
	Service    string
	Dependency string
}

type AppPaths struct {
	Attorney                                                   AttorneyPaths
	CertificateProvider                                        CertificateProviderPaths
	AboutPayment                                               string
	AreYouHappyIfOneAttorneyCantActNoneCan                     string
	AreYouHappyIfOneReplacementAttorneyCantActNoneCan          string
	AreYouHappyIfRemainingAttorneysCanContinueToAct            string
	AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct string
	AuthRedirect                                               string
	CertificateProviderAddress                                 string
	CertificateProviderDetails                                 string
	CertificateProviderOptOut                                  string
	CertificateProviderStart                                   string
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
	HealthCheck                                                HealthCheckPaths
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
	SignOut                                                    string
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
		CheckYourName:             "/attorney-check-your-name",
		CodeOfConduct:             "/attorney-code-of-conduct",
		EnterReferenceNumber:      "/attorney-enter-reference-number",
		Login:                     "/attorney-login",
		LoginCallback:             "/attorney-login-callback",
		MobileNumber:              "/attorney-mobile-number",
		ReadTheLpa:                "/attorney-read-the-lpa",
		RightsAndResponsibilities: "/attorney-legal-rights-and-responsibilities",
		Sign:                      "/attorney-sign",
		Start:                     "/attorney-start",
		TaskList:                  "/attorney-task-list",
		WhatHappensNext:           "/attorney-what-happens-next",
		WhatHappensWhenYouSign:    "/attorney-what-happens-when-you-sign-the-lpa",
	},
	AuthRedirect:               "/auth/redirect",
	CertificateProviderDetails: "/certificate-provider-details",
	CertificateProvider: CertificateProviderPaths{
		CertificateProvided:                  "/certificate-provided",
		CheckYourName:                        "/certificate-provider-check-your-name",
		EnterDateOfBirth:                     "/certificate-provider-enter-date-of-birth",
		EnterMobileNumber:                    "/certificate-provider-enter-mobile-number",
		EnterReferenceNumber:                 "/certificate-provider-enter-reference-number",
		IdentityWithBiometricResidencePermit: "/certificate-provider/id/brp",
		IdentityWithDrivingLicencePaper:      "/certificate-provider/id/dlpaper",
		IdentityWithDrivingLicencePhotocard:  "/certificate-provider/id/dlphoto",
		IdentityWithOneLogin:                 "/certificate-provider-identity-with-one-login",
		IdentityWithOneLoginCallback:         "/certificate-provider-identity-with-one-login-callback",
		IdentityWithOnlineBankAccount:        "/certificate-provider/id/bank",
		IdentityWithPassport:                 "/certificate-provider/id/passport",
		IdentityWithYoti:                     "/certificate-provider-identity-with-yoti",
		IdentityWithYotiCallback:             "/certificate-provider-identity-with-yoti-callback",
		Login:                                "/certificate-provider-login",
		LoginCallback:                        "/certificate-provider-login-callback",
		ProvideCertificate:                   "/provide-certificate",
		ReadTheLpa:                           "/certificate-provider-read-the-lpa",
		SelectYourIdentityOptions1:           "/certificate-provider-select-identity-document",
		SelectYourIdentityOptions2:           "/certificate-provider-select-identity-document-2",
		SelectYourIdentityOptions:            "/certificate-provider-select-your-identity-options",
		WhatHappensNext:                      "/certificate-provider-what-happens-next",
		WhatYoullNeedToConfirmYourIdentity:   "/certificate-provider-what-youll-need-to-confirm-your-identity",
		WhoIsEligible:                        "/certificate-provider-who-is-eligible",
		YourChosenIdentityOptions:            "/certificate-provider-your-chosen-identity-options",
	},
	CertificateProviderOptOut:         "/certificate-provider-opt-out",
	CertificateProviderAddress:        "/certificate-provider-address",
	CertificateProviderStart:          "/certificate-provider-start",
	CheckYourLpa:                      "/check-your-lpa",
	ChooseAttorneys:                   "/choose-attorneys",
	ChooseAttorneysAddress:            "/choose-attorneys-address",
	ChooseAttorneysSummary:            "/choose-attorneys-summary",
	ChoosePeopleToNotify:              "/choose-people-to-notify",
	ChoosePeopleToNotifyAddress:       "/choose-people-to-notify-address",
	ChoosePeopleToNotifySummary:       "/choose-people-to-notify-summary",
	ChooseReplacementAttorneys:        "/choose-replacement-attorneys",
	ChooseReplacementAttorneysAddress: "/choose-replacement-attorneys-address",
	ChooseReplacementAttorneysSummary: "/choose-replacement-attorneys-summary",
	CookiesConsent:                    "/cookies-consent",
	Dashboard:                         "/dashboard",
	DoYouWantReplacementAttorneys:     "/do-you-want-replacement-attorneys",
	DoYouWantToNotifyPeople:           "/do-you-want-to-notify-people",
	Fixtures:                          "/fixtures",
	HealthCheck: HealthCheckPaths{
		Service:    "/health-check/service",
		Dependency: "/health-check/dependency",
	},
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
	SignOut:                                              "/sign-out",
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

// IsAttorneyPath checks whether the url should be prefixed with /attorney/.
func IsAttorneyPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return slices.Contains([]string{
		Paths.Attorney.CheckYourName,
		Paths.Attorney.CodeOfConduct,
		Paths.Attorney.MobileNumber,
		Paths.Attorney.ReadTheLpa,
		Paths.Attorney.RightsAndResponsibilities,
		Paths.Attorney.Sign,
		Paths.Attorney.TaskList,
		Paths.Attorney.WhatHappensNext,
		Paths.Attorney.WhatHappensWhenYouSign,
	}, path)
}

// IsCertificateProviderPath checks whether the url should be prefixed with /certificate-provider/.
func IsCertificateProviderPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return slices.Contains([]string{
		Paths.CertificateProvider.CertificateProvided,
		Paths.CertificateProvider.CheckYourName,
		Paths.CertificateProvider.EnterDateOfBirth,
		Paths.CertificateProvider.EnterMobileNumber,
		Paths.CertificateProvider.IdentityWithBiometricResidencePermit,
		Paths.CertificateProvider.IdentityWithDrivingLicencePaper,
		Paths.CertificateProvider.IdentityWithDrivingLicencePhotocard,
		Paths.CertificateProvider.IdentityWithOneLogin,
		Paths.CertificateProvider.IdentityWithOneLoginCallback,
		Paths.CertificateProvider.IdentityWithOnlineBankAccount,
		Paths.CertificateProvider.IdentityWithPassport,
		Paths.CertificateProvider.IdentityWithYoti,
		Paths.CertificateProvider.IdentityWithYotiCallback,
		Paths.CertificateProvider.ProvideCertificate,
		Paths.CertificateProvider.ReadTheLpa,
		Paths.CertificateProvider.SelectYourIdentityOptions,
		Paths.CertificateProvider.SelectYourIdentityOptions1,
		Paths.CertificateProvider.SelectYourIdentityOptions2,
		Paths.CertificateProvider.WhatHappensNext,
		Paths.CertificateProvider.WhatYoullNeedToConfirmYourIdentity,
		Paths.CertificateProvider.YourChosenIdentityOptions,
	}, path)
}

// IsLpaPath checks whether the url should be prefixed with /lpa/.
func IsLpaPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return !IsAttorneyPath(url) &&
		!IsCertificateProviderPath(url) &&
		!slices.Contains([]string{
			Paths.Attorney.EnterReferenceNumber,
			Paths.Attorney.Login,
			Paths.Attorney.LoginCallback,
			Paths.Attorney.Start,
			Paths.AuthRedirect,
			Paths.CertificateProvider.EnterReferenceNumber,
			Paths.CertificateProvider.Login,
			Paths.CertificateProvider.LoginCallback,
			Paths.CertificateProvider.WhoIsEligible,
			Paths.CertificateProviderStart,
			Paths.Dashboard,
			Paths.Login,
			Paths.LoginCallback,
			Paths.SignOut,
			Paths.Start,
			Paths.YotiRedirect,
		}, path)
}
