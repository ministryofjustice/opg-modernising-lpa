package page

import "strings"

type AppPaths struct {
	AboutPayment                                         string
	Auth                                                 string
	AuthRedirect                                         string
	CertificateProviderAddress                           string
	CertificateProviderDetails                           string
	CertificateProviderLogin                             string
	CertificateProviderLoginCallback                     string
	CertificateProviderStart                             string
	CertificateProviderYourDetails                       string
	CheckYourLpa                                         string
	ChooseAttorneys                                      string
	ChooseAttorneysAddress                               string
	ChooseAttorneysSummary                               string
	ChoosePeopleToNotify                                 string
	ChoosePeopleToNotifyAddress                          string
	ChoosePeopleToNotifySummary                          string
	ChooseReplacementAttorneys                           string
	ChooseReplacementAttorneysAddress                    string
	ChooseReplacementAttorneysSummary                    string
	CookiesConsent                                       string
	Dashboard                                            string
	DoYouWantReplacementAttorneys                        string
	DoYouWantToNotifyPeople                              string
	HealthCheck                                          string
	HowDoYouKnowYourCertificateProvider                  string
	HowLongHaveYouKnownCertificateProvider               string
	HowShouldAttorneysMakeDecisions                      string
	HowShouldReplacementAttorneysMakeDecisions           string
	HowShouldReplacementAttorneysStepIn                  string
	HowToConfirmYourIdentityAndSign                      string
	HowWouldCertificateProviderPreferToCarryOutTheirRole string
	IdentityConfirmed                                    string
	IdentityWithBiometricResidencePermit                 string
	IdentityWithDrivingLicencePaper                      string
	IdentityWithDrivingLicencePhotocard                  string
	IdentityWithOneLogin                                 string
	IdentityWithOneLoginCallback                         string
	IdentityWithOnlineBankAccount                        string
	IdentityWithPassport                                 string
	IdentityWithYoti                                     string
	IdentityWithYotiCallback                             string
	LpaType                                              string
	PaymentConfirmation                                  string
	Progress                                             string
	ReadYourLpa                                          string
	RemoveAttorney                                       string
	RemovePersonToNotify                                 string
	RemoveReplacementAttorney                            string
	Restrictions                                         string
	Root                                                 string
	SelectYourIdentityOptions                            string
	SelectYourIdentityOptions1                           string
	SelectYourIdentityOptions2                           string
	SignYourLpa                                          string
	Start                                                string
	TaskList                                             string
	TestingStart                                         string
	WhatYoullNeedToConfirmYourIdentity                   string
	WhenCanTheLpaBeUsed                                  string
	WhoDoYouWantToBeCertificateProviderGuidance          string
	WhoIsTheLpaFor                                       string
	WitnessingAsCertificateProvider                      string
	WitnessingYourSignature                              string
	YouHaveSubmittedYourLpa                              string
	YourAddress                                          string
	YourChosenIdentityOptions                            string
	YourDetails                                          string
	YourLegalRightsAndResponsibilities                   string
}

var Paths = AppPaths{
	AboutPayment:                                         "/about-payment",
	Auth:                                                 "/auth",
	AuthRedirect:                                         "/auth/redirect",
	CertificateProviderAddress:                           "/certificate-provider-address",
	CertificateProviderDetails:                           "/certificate-provider-details",
	CertificateProviderLogin:                             "/certificate-provider-login",
	CertificateProviderLoginCallback:                     "/certificate-provider-login-callback",
	CertificateProviderStart:                             "/certificate-provider-start",
	CertificateProviderYourDetails:                       "/certificate-provider-your-details",
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
	LpaType:                                              "/lpa-type",
	PaymentConfirmation:                                  "/payment-confirmation",
	Progress:                                             "/progress",
	ReadYourLpa:                                          "/read-your-lpa",
	RemoveAttorney:                                       "/remove-attorney",
	RemovePersonToNotify:                                 "/remove-person-to-notify",
	RemoveReplacementAttorney:                            "/remove-replacement-attorney",
	Restrictions:                                         "/restrictions",
	Root:                                                 "/",
	SelectYourIdentityOptions1:                           "/select-your-identity-options-1",
	SelectYourIdentityOptions2:                           "/select-your-identity-options-2",
	SelectYourIdentityOptions:                            "/select-your-identity-options",
	SignYourLpa:                                          "/sign-your-lpa",
	Start:                                                "/start",
	TaskList:                                             "/task-list",
	TestingStart:                                         "/testing-start",
	WhatYoullNeedToConfirmYourIdentity:                   "/what-youll-need-to-confirm-your-identity",
	WhenCanTheLpaBeUsed:                                  "/when-can-the-lpa-be-used",
	WhoDoYouWantToBeCertificateProviderGuidance:          "/who-do-you-want-to-be-certificate-provider-guidance",
	WhoIsTheLpaFor:                                       "/who-is-the-lpa-for",
	WitnessingAsCertificateProvider:                      "/witnessing-as-certificate-provider",
	WitnessingYourSignature:                              "/witnessing-your-signature",
	YouHaveSubmittedYourLpa:                              "/you-have-submitted-your-lpa",
	YourAddress:                                          "/your-address",
	YourChosenIdentityOptions:                            "/your-chosen-identity-options",
	YourDetails:                                          "/your-details",
	YourLegalRightsAndResponsibilities:                   "/your-legal-rights-and-responsibilities",
}

func IsLpaPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return path != Paths.Auth && path != Paths.AuthRedirect &&
		path != Paths.Dashboard && path != Paths.Start &&
		path != Paths.CertificateProviderLogin && path != Paths.CertificateProviderLoginCallback && path != Paths.CertificateProviderYourDetails
}
