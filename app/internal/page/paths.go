package page

import (
	"strings"

	"golang.org/x/exp/slices"
)

type AppPaths struct {
	AboutPayment                                            string
	AuthRedirect                                            string
	CertificateProvided                                     string
	CertificateProviderAddress                              string
	CertificateProviderDetails                              string
	CertificateProviderEnterReferenceNumber                 string
	CertificateProviderGuidance                             string
	CertificateProviderIdentityWithBiometricResidencePermit string
	CertificateProviderIdentityWithDrivingLicencePaper      string
	CertificateProviderIdentityWithDrivingLicencePhotocard  string
	CertificateProviderIdentityWithOneLogin                 string
	CertificateProviderIdentityWithOneLoginCallback         string
	CertificateProviderIdentityWithOnlineBankAccount        string
	CertificateProviderIdentityWithPassport                 string
	CertificateProviderIdentityWithYoti                     string
	CertificateProviderIdentityWithYotiCallback             string
	CertificateProviderLogin                                string
	CertificateProviderLoginCallback                        string
	CertificateProviderOptOut                               string
	CertificateProviderReadTheLpa                           string
	CertificateProviderSelectYourIdentityOptions            string
	CertificateProviderSelectYourIdentityOptions1           string
	CertificateProviderSelectYourIdentityOptions2           string
	CertificateProviderStart                                string
	CertificateProviderWhatHappensNext                      string
	CertificateProviderWhatYoullNeedToConfirmYourIdentity   string
	CertificateProviderYourAddress                          string
	CertificateProviderYourChosenIdentityOptions            string
	CertificateProviderYourDetails                          string
	CheckYourLpa                                            string
	ChooseAttorneys                                         string
	ChooseAttorneysAddress                                  string
	ChooseAttorneysSummary                                  string
	ChoosePeopleToNotify                                    string
	ChoosePeopleToNotifyAddress                             string
	ChoosePeopleToNotifySummary                             string
	ChooseReplacementAttorneys                              string
	ChooseReplacementAttorneysAddress                       string
	ChooseReplacementAttorneysSummary                       string
	CookiesConsent                                          string
	Dashboard                                               string
	DoYouWantReplacementAttorneys                           string
	DoYouWantToNotifyPeople                                 string
	Fixtures                                                string
	HealthCheck                                             string
	HowDoYouKnowTheDonor                                    string
	HowDoYouKnowYourCertificateProvider                     string
	HowLongHaveYouKnownCertificateProvider                  string
	HowLongHaveYouKnownDonor                                string
	HowShouldAttorneysMakeDecisions                         string
	HowShouldReplacementAttorneysMakeDecisions              string
	HowShouldReplacementAttorneysStepIn                     string
	HowToConfirmYourIdentityAndSign                         string
	HowWouldCertificateProviderPreferToCarryOutTheirRole    string
	IdentityConfirmed                                       string
	IdentityWithBiometricResidencePermit                    string
	IdentityWithDrivingLicencePaper                         string
	IdentityWithDrivingLicencePhotocard                     string
	IdentityWithOneLogin                                    string
	IdentityWithOneLoginCallback                            string
	IdentityWithOnlineBankAccount                           string
	IdentityWithPassport                                    string
	IdentityWithYoti                                        string
	IdentityWithYotiCallback                                string
	Login                                                   string
	LoginCallback                                           string
	LpaType                                                 string
	PaymentConfirmation                                     string
	Progress                                                string
	ProvideCertificate                                      string
	ReadYourLpa                                             string
	RemoveAttorney                                          string
	RemovePersonToNotify                                    string
	RemoveReplacementAttorney                               string
	Restrictions                                            string
	Root                                                    string
	SelectYourIdentityOptions                               string
	SelectYourIdentityOptions1                              string
	SelectYourIdentityOptions2                              string
	SignYourLpa                                             string
	Start                                                   string
	TaskList                                                string
	TestingStart                                            string
	WhatYoullNeedToConfirmYourIdentity                      string
	WhenCanTheLpaBeUsed                                     string
	WhoDoYouWantToBeCertificateProviderGuidance             string
	WhoIsTheLpaFor                                          string
	WitnessingAsCertificateProvider                         string
	WitnessingYourSignature                                 string
	YouHaveSubmittedYourLpa                                 string
	YourAddress                                             string
	YourChosenIdentityOptions                               string
	YourDetails                                             string
	YourLegalRightsAndResponsibilities                      string
}

var Paths = AppPaths{
	AboutPayment:                            "/about-payment",
	AuthRedirect:                            "/auth/redirect",
	CertificateProvided:                     "/certificate-provided",
	CertificateProviderAddress:              "/certificate-provider-address",
	CertificateProviderDetails:              "/certificate-provider-details",
	CertificateProviderEnterReferenceNumber: "/certificate-provider-enter-reference-number",
	CertificateProviderGuidance:             "/being-a-certificate-provider",
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
	CertificateProviderSelectYourIdentityOptions1:           "/certificate-provider-select-your-identity-options-1",
	CertificateProviderSelectYourIdentityOptions2:           "/certificate-provider-select-your-identity-options-2",
	CertificateProviderSelectYourIdentityOptions:            "/certificate-provider-select-your-identity-options",
	CertificateProviderStart:                                "/certificate-provider-start",
	CertificateProviderWhatHappensNext:                      "/certificate-provider-what-happens-next",
	CertificateProviderWhatYoullNeedToConfirmYourIdentity:   "/certificate-provider-what-youll-need-to-confirm-your-identity",
	CertificateProviderYourAddress:                          "/certificate-provider-your-address",
	CertificateProviderYourChosenIdentityOptions:            "/certificate-provider-your-chosen-identity-options",
	CertificateProviderYourDetails:                          "/certificate-provider-your-details",
	CheckYourLpa:                                            "/check-your-lpa",
	ChooseAttorneys:                                         "/choose-attorneys",
	ChooseAttorneysAddress:                                  "/choose-attorneys-address",
	ChooseAttorneysSummary:                                  "/choose-attorneys-summary",
	ChoosePeopleToNotify:                                    "/choose-people-to-notify",
	ChoosePeopleToNotifyAddress:                             "/choose-people-to-notify-address",
	ChoosePeopleToNotifySummary:                             "/choose-people-to-notify-summary",
	ChooseReplacementAttorneys:                              "/choose-replacement-attorneys",
	ChooseReplacementAttorneysAddress:                       "/choose-replacement-attorneys-address",
	ChooseReplacementAttorneysSummary:                       "/choose-replacement-attorneys-summary",
	CookiesConsent:                                          "/cookies-consent",
	Dashboard:                                               "/dashboard",
	DoYouWantReplacementAttorneys:                           "/do-you-want-replacement-attorneys",
	DoYouWantToNotifyPeople:                                 "/do-you-want-to-notify-people",
	Fixtures:                                                "/fixtures",
	HealthCheck:                                             "/health-check",
	HowDoYouKnowTheDonor:                                    "/how-do-you-know-the-donor",
	HowDoYouKnowYourCertificateProvider:                     "/how-do-you-know-your-certificate-provider",
	HowLongHaveYouKnownCertificateProvider:                  "/how-long-have-you-known-certificate-provider",
	HowLongHaveYouKnownDonor:                                "/how-long-have-you-known-donor",
	HowShouldAttorneysMakeDecisions:                         "/how-should-attorneys-make-decisions",
	HowShouldReplacementAttorneysMakeDecisions:              "/how-should-replacement-attorneys-make-decisions",
	HowShouldReplacementAttorneysStepIn:                     "/how-should-replacement-attorneys-step-in",
	HowToConfirmYourIdentityAndSign:                         "/how-to-confirm-your-identity-and-sign",
	HowWouldCertificateProviderPreferToCarryOutTheirRole:    "/how-would-certificate-provider-prefer-to-carry-out-their-role",
	IdentityConfirmed:                                       "/identity-confirmed",
	IdentityWithBiometricResidencePermit:                    "/id/biometric-residence-permit",
	IdentityWithDrivingLicencePaper:                         "/id/driving-licence-paper",
	IdentityWithDrivingLicencePhotocard:                     "/id/driving-licence-photocard",
	IdentityWithOneLogin:                                    "/id/one-login",
	IdentityWithOneLoginCallback:                            "/id/one-login/callback",
	IdentityWithOnlineBankAccount:                           "/id/online-bank-account",
	IdentityWithPassport:                                    "/id/passport",
	IdentityWithYoti:                                        "/id/yoti",
	IdentityWithYotiCallback:                                "/id/yoti/callback",
	Login:                                                   "/login",
	LoginCallback:                                           "/login-callback",
	LpaType:                                                 "/lpa-type",
	PaymentConfirmation:                                     "/payment-confirmation",
	Progress:                                                "/progress",
	ProvideCertificate:                                      "/provide-certificate",
	ReadYourLpa:                                             "/read-your-lpa",
	RemoveAttorney:                                          "/remove-attorney",
	RemovePersonToNotify:                                    "/remove-person-to-notify",
	RemoveReplacementAttorney:                               "/remove-replacement-attorney",
	Restrictions:                                            "/restrictions",
	Root:                                                    "/",
	SelectYourIdentityOptions1:                              "/select-your-identity-options-1",
	SelectYourIdentityOptions2:                              "/select-your-identity-options-2",
	SelectYourIdentityOptions:                               "/select-your-identity-options",
	SignYourLpa:                                             "/sign-your-lpa",
	Start:                                                   "/start",
	TaskList:                                                "/task-list",
	TestingStart:                                            "/testing-start",
	WhatYoullNeedToConfirmYourIdentity:                      "/what-youll-need-to-confirm-your-identity",
	WhenCanTheLpaBeUsed:                                     "/when-can-the-lpa-be-used",
	WhoDoYouWantToBeCertificateProviderGuidance:             "/who-do-you-want-to-be-certificate-provider-guidance",
	WhoIsTheLpaFor:                                          "/who-is-the-lpa-for",
	WitnessingAsCertificateProvider:                         "/witnessing-as-certificate-provider",
	WitnessingYourSignature:                                 "/witnessing-your-signature",
	YouHaveSubmittedYourLpa:                                 "/you-have-submitted-your-lpa",
	YourAddress:                                             "/your-address",
	YourChosenIdentityOptions:                               "/your-chosen-identity-options",
	YourDetails:                                             "/your-details",
	YourLegalRightsAndResponsibilities:                      "/your-legal-rights-and-responsibilities",
}

func IsLpaPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return !slices.Contains([]string{
		Paths.AuthRedirect,
		Paths.CertificateProvided,
		Paths.CertificateProviderEnterReferenceNumber,
		Paths.CertificateProviderGuidance,
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
		Paths.CertificateProviderYourAddress,
		Paths.CertificateProviderYourChosenIdentityOptions,
		Paths.CertificateProviderYourDetails,
		Paths.Dashboard,
		Paths.HowDoYouKnowTheDonor,
		Paths.HowLongHaveYouKnownDonor,
		Paths.Login,
		Paths.LoginCallback,
		Paths.ProvideCertificate,
		Paths.Start,
	}, path)
}
