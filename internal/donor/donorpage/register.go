// Package donorpage provides the pages that a donor interacts with.
package donorpage

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpadata.Lpa, error)
}

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error

type Template func(io.Writer, interface{}) error

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type DonorStore interface {
	Get(ctx context.Context) (*donordata.Provided, error)
	Latest(ctx context.Context) (*donordata.Provided, error)
	Put(ctx context.Context, donor *donordata.Provided) error
	Delete(ctx context.Context) error
	Link(ctx context.Context, data sharecodedata.Link, donorEmail string) error
	DeleteVoucher(ctx context.Context, provided *donordata.Provided) error
}

type GetDonorStore interface {
	Get(context.Context) (*donordata.Provided, error)
}

type ReuseStore interface {
	PutCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error
	Correspondents(ctx context.Context) ([]donordata.Correspondent, error)
	DeleteCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error
	PutAttorney(ctx context.Context, attorney donordata.Attorney) error
	PutAttorneys(ctx context.Context, attorneys []donordata.Attorney) error
	Attorneys(ctx context.Context, provided *donordata.Provided) ([]donordata.Attorney, error)
	DeleteAttorney(ctx context.Context, attorney donordata.Attorney) error
	PutTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error
	TrustCorporations(ctx context.Context) ([]donordata.TrustCorporation, error)
	DeleteTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error
	PutCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	CertificateProviders(ctx context.Context) ([]donordata.CertificateProvider, error)
	DeleteCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error
	PutPersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
	PutPeopleToNotify(ctx context.Context, peopleToNotify []donordata.PersonToNotify) error
	PeopleToNotify(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error)
	DeletePersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*certificateproviderdata.Provided, error)
}

type EvidenceReceivedStore interface {
	Get(context.Context) (bool, error)
}

type S3Client interface {
	PutObject(context.Context, string, []byte) error
	DeleteObject(context.Context, string) error
}

type PayClient interface {
	CreatePayment(ctx context.Context, lpaUID string, body pay.CreatePaymentBody) (*pay.CreatePaymentResponse, error)
	GetPayment(ctx context.Context, id string) (pay.GetPaymentResponse, error)
	CanRedirect(url string) bool
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type ShareCodeSender interface {
	SendCertificateProviderInvite(ctx context.Context, appData appcontext.Data, provided *donordata.Provided) error
	SendCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, provided *donordata.Provided) error
	SendVoucherAccessCode(ctx context.Context, donor *donordata.Provided, appData appcontext.Data) error
	SendVoucherInvite(ctx context.Context, donor *donordata.Provided, appData appcontext.Data) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, confidenceLevel onelogin.ConfidenceLevel) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(userInfo onelogin.UserInfo) (identity.UserData, error)
}

type NotifyClient interface {
	SendActorSMS(ctx context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error
	SendEmail(ctx context.Context, to notify.ToEmail, email notify.Email) error
	SendActorEmail(ctx context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
}

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
	SetPayment(r *http.Request, w http.ResponseWriter, session *sesh.PaymentSession) error
	Payment(r *http.Request) (*sesh.PaymentSession, error)
	ClearPayment(r *http.Request, w http.ResponseWriter) error
}

type WitnessCodeSender interface {
	SendToCertificateProvider(context.Context, *donordata.Provided) error
	SendToIndependentWitness(context.Context, *donordata.Provided) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

type RequestSigner interface {
	Sign(context.Context, *http.Request, string) error
}

type Localizer interface {
	localize.Localizer
}

type DocumentStore interface {
	GetAll(context.Context) (document.Documents, error)
	Put(context.Context, document.Document) error
	Delete(context.Context, document.Document) error
	DeleteInfectedDocuments(context.Context, document.Documents) error
	Create(context.Context, *donordata.Provided, string, []byte) (document.Document, error)
	Submit(context.Context, *donordata.Provided, document.Documents) error
}

type EventClient interface {
	SendReducedFeeRequested(ctx context.Context, e event.ReducedFeeRequested) error
	SendPaymentReceived(ctx context.Context, e event.PaymentReceived) error
	SendUidRequested(ctx context.Context, e event.UidRequested) error
	SendCertificateProviderStarted(ctx context.Context, e event.CertificateProviderStarted) error
	SendIdentityCheckMismatched(ctx context.Context, e event.IdentityCheckMismatched) error
	SendCorrespondentUpdated(ctx context.Context, e event.CorrespondentUpdated) error
	SendConfirmAtPostOfficeSelected(ctx context.Context, e event.ConfirmAtPostOfficeSelected) error
	SendRegisterWithCourtOfProtection(ctx context.Context, e event.RegisterWithCourtOfProtection) error
}

type DashboardStore interface {
	GetAll(ctx context.Context) (results dashboarddata.Results, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	Lpa(ctx context.Context, lpaUID string) (*lpadata.Lpa, error)
	SendDonorConfirmIdentity(ctx context.Context, donor *donordata.Provided) error
	SendLpa(ctx context.Context, uid string, body lpastore.CreateLpa) error
	SendDonorWithdrawLPA(ctx context.Context, lpaUID string) error
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, code sharecodedata.Hashed) (sharecodedata.Link, error)
}

type ScheduledStore interface {
	Create(ctx context.Context, rows ...scheduled.Event) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type ProgressTracker interface {
	Progress(lpa *lpadata.Lpa) task.Progress
}

type VoucherStore interface {
	Get(ctx context.Context) (*voucherdata.Provided, error)
	GetAny(ctx context.Context) (*voucherdata.Provided, error)
}

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

type PeopleToNotifyService interface {
	Reusable(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error)
	WantPeopleToNotify(ctx context.Context, provided *donordata.Provided, yesNo form.YesNo) error
	PutMany(ctx context.Context, provided *donordata.Provided, people []donordata.PersonToNotify) error
	Put(ctx context.Context, provided *donordata.Provided, person donordata.PersonToNotify) (actoruid.UID, error)
	Delete(ctx context.Context, provided *donordata.Provided, person donordata.PersonToNotify) error
}

type CertificateProviderService interface {
	Reusable(ctx context.Context) ([]donordata.CertificateProvider, error)
	Put(ctx context.Context, provided *donordata.Provided) error
	Delete(ctx context.Context, provided *donordata.Provided) error
}

type CorrespondentService interface {
	Reusable(ctx context.Context) ([]donordata.Correspondent, error)
	NotWanted(ctx context.Context, provided *donordata.Provided) error
	Put(ctx context.Context, provided *donordata.Provided) error
	Delete(ctx context.Context, provided *donordata.Provided) error
}

type AttorneyService interface {
	Reusable(ctx context.Context, provided *donordata.Provided) ([]donordata.Attorney, error)
	ReusableTrustCorporations(ctx context.Context) ([]donordata.TrustCorporation, error)
	PutMany(ctx context.Context, provided *donordata.Provided, attorneys []donordata.Attorney) error
	Put(ctx context.Context, provided *donordata.Provided, attorney donordata.Attorney) error
	PutTrustCorporation(ctx context.Context, provided *donordata.Provided, trustCorporation donordata.TrustCorporation) error
	Delete(ctx context.Context, provided *donordata.Provided, attorney donordata.Attorney) error
	DeleteTrustCorporation(ctx context.Context, provided *donordata.Provided) error
	WantReplacements(ctx context.Context, provided *donordata.Provided, yesNo form.YesNo) error
	IsReplacement() bool
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	donorStore DonorStore,
	oneLoginClient OneLoginClient,
	addressClient AddressClient,
	appPublicURL string,
	payClient PayClient,
	shareCodeSender ShareCodeSender,
	witnessCodeSender WitnessCodeSender,
	errorHandler page.ErrorHandler,
	certificateProviderStore CertificateProviderStore,
	notifyClient NotifyClient,
	evidenceReceivedStore EvidenceReceivedStore,
	documentStore DocumentStore,
	eventClient EventClient,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
	shareCodeStore ShareCodeStore,
	progressTracker ProgressTracker,
	lpaStoreResolvingService LpaStoreResolvingService,
	scheduledStore ScheduledStore,
	voucherStore VoucherStore,
	reuseStore ReuseStore,
	bundle Bundle,
	donorStartURL string,
	certificateProviderStartURL string,
	attorneyStartURL string,
) {
	payer := Pay(logger, sessionStore, donorStore, payClient, appPublicURL)

	handleRoot := makeHandle(rootMux, sessionStore, errorHandler, donorStartURL)

	handleRoot(page.PathLogin, page.None,
		page.Login(oneLoginClient, sessionStore, random.String, page.PathLoginCallback))
	handleRoot(page.PathLoginCallback, page.None,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.PathMakeOrAddAnLPA, dashboardStore, actor.TypeDonor))
	handleRoot(page.PathEnterAccessCode, page.RequireSession,
		EnterAccessCode(logger, tmpls.Get("enter_access_code.gohtml"), shareCodeStore, donorStore, sessionStore))

	handleWithDonor := makeLpaHandle(rootMux, sessionStore, errorHandler, donorStore, donorStartURL)

	handleWithDonor(donor.PathViewLPA, page.None,
		ViewLpa(tmpls.Get("view_lpa.gohtml"), lpaStoreClient))

	handleWithDonor(donor.PathDeleteThisLpa, page.None,
		DeleteLpa(tmpls.Get("delete_this_lpa.gohtml"), donorStore, notifyClient, certificateProviderStartURL))
	handleWithDonor(donor.PathWithdrawThisLpa, page.None,
		WithdrawLpa(tmpls.Get("withdraw_this_lpa.gohtml"), donorStore, time.Now, lpaStoreClient, notifyClient, lpaStoreResolvingService, certificateProviderStore, certificateProviderStartURL, attorneyStartURL))

	handleWithDonor(donor.PathMakeANewLPA, page.None,
		Guidance(tmpls.Get("make_a_new_lpa.gohtml")))
	handleWithDonor(donor.PathYourDetails, page.None,
		Guidance(tmpls.Get("your_details.gohtml")))
	handleWithDonor(donor.PathYourName, page.None,
		YourName(tmpls.Get("your_name.gohtml"), donorStore, sessionStore))
	handleWithDonor(donor.PathYourDateOfBirth, page.CanGoBack,
		YourDateOfBirth(tmpls.Get("your_date_of_birth.gohtml"), donorStore))
	handleWithDonor(donor.PathYouHaveToldUsYouAreUnder18, page.CanGoBack,
		Guidance(tmpls.Get("you_have_told_us_you_are_under_18.gohtml")))
	handleWithDonor(donor.PathDoYouLiveInTheUK, page.CanGoBack,
		DoYouLiveInTheUK(tmpls.Get("do_you_live_in_the_uk.gohtml")))
	handleWithDonor(donor.PathYourAddress, page.CanGoBack,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, donorStore))
	handleWithDonor(donor.PathWhatCountryDoYouLiveIn, page.CanGoBack,
		WhatCountryDoYouLiveIn(tmpls.Get("what_country_do_you_live_in.gohtml"), donorStore))
	handleWithDonor(donor.PathYourNonUKAddress, page.CanGoBack,
		YourNonUKAddress(tmpls.Get("your_non_uk_address.gohtml"), donorStore))
	handleWithDonor(donor.PathReceivingUpdatesAboutYourLpa, page.CanGoBack,
		Guidance(tmpls.Get("receiving_updates_about_your_lpa.gohtml")))
	handleWithDonor(donor.PathYourEmail, page.CanGoBack,
		YourEmail(tmpls.Get("your_email.gohtml"), donorStore, nil))
	handleWithDonor(donor.PathYourMobile, page.CanGoBack,
		YourMobile(tmpls.Get("your_mobile.gohtml"), donorStore, nil))
	handleWithDonor(donor.PathWeHaveUpdatedYourDetails, page.None,
		Guidance(tmpls.Get("we_have_updated_your_details.gohtml")))
	handleWithDonor(donor.PathCanYouSignYourLpa, page.CanGoBack,
		CanYouSignYourLpa(tmpls.Get("can_you_sign_your_lpa.gohtml"), donorStore))
	handleWithDonor(donor.PathCheckYouCanSign, page.CanGoBack,
		CheckYouCanSign(tmpls.Get("check_you_can_sign.gohtml"), donorStore))
	handleWithDonor(donor.PathYourPreferredLanguage, page.CanGoBack,
		YourPreferredLanguage(tmpls.Get("your_preferred_language.gohtml"), donorStore))
	handleWithDonor(donor.PathYourLegalRightsAndResponsibilitiesIfYouMakeLpa, page.CanGoBack,
		Guidance(tmpls.Get("your_legal_rights_and_responsibilities_if_you_make_lpa.gohtml")))
	handleWithDonor(donor.PathLpaType, page.CanGoBack,
		LpaType(tmpls.Get("lpa_type.gohtml"), donorStore, eventClient, sessionStore))
	handleWithDonor(donor.PathNeedHelpSigningConfirmation, page.None,
		Guidance(tmpls.Get("need_help_signing_confirmation.gohtml")))

	handleWithDonor(donor.PathTaskList, page.None,
		TaskList(tmpls.Get("task_list.gohtml"), evidenceReceivedStore))

	{
		service := donor.NewAttorneyService(donorStore, reuseStore)

		handleWithDonor(donor.PathChooseAttorneysGuidance, page.None,
			Guidance(tmpls.Get("choose_attorneys_guidance.gohtml")))
		handleWithDonor(donor.PathChooseAttorneys, page.CanGoBack,
			ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterAttorney, page.CanGoBack,
			EnterAttorney(tmpls.Get("enter_attorney.gohtml"), service))
		handleWithDonor(donor.PathChooseAttorneysAddress, page.CanGoBack,
			ChooseAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathChooseTrustCorporation, page.CanGoBack,
			ChooseTrustCorporation(tmpls.Get("choose_trust_corporation.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterTrustCorporation, page.CanGoBack,
			EnterTrustCorporation(tmpls.Get("enter_trust_corporation.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterTrustCorporationAddress, page.CanGoBack,
			EnterTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathChooseAttorneysSummary, page.CanGoBack,
			ChooseAttorneysSummary(tmpls.Get("choose_attorneys_summary.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathRemoveAttorney, page.CanGoBack,
			RemoveAttorney(tmpls.Get("remove_attorney.gohtml"), service))
		handleWithDonor(donor.PathRemoveTrustCorporation, page.CanGoBack,
			RemoveTrustCorporation(tmpls.Get("remove_attorney.gohtml"), service))
		handleWithDonor(donor.PathHowShouldAttorneysMakeDecisions, page.CanGoBack,
			HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), donorStore))
		handleWithDonor(donor.PathBecauseYouHaveChosenJointly, page.CanGoBack,
			Guidance(tmpls.Get("because_you_have_chosen_jointly.gohtml")))
		handleWithDonor(donor.PathBecauseYouHaveChosenJointlyForSomeSeverallyForOthers, page.CanGoBack,
			Guidance(tmpls.Get("because_you_have_chosen_jointly_for_some_severally_for_others.gohtml")))
	}

	{
		service := donor.NewReplacementAttorneyService(donorStore, reuseStore)

		handleWithDonor(donor.PathDoYouWantReplacementAttorneys, page.None,
			WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), service))
		handleWithDonor(donor.PathChooseReplacementAttorneys, page.CanGoBack,
			ChooseAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterReplacementAttorney, page.CanGoBack,
			EnterAttorney(tmpls.Get("enter_replacement_attorney.gohtml"), service))
		handleWithDonor(donor.PathChooseReplacementAttorneysAddress, page.CanGoBack,
			ChooseAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathChooseReplacementTrustCorporation, page.CanGoBack,
			ChooseTrustCorporation(tmpls.Get("choose_trust_corporation.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterReplacementTrustCorporation, page.CanGoBack,
			EnterTrustCorporation(tmpls.Get("enter_replacement_trust_corporation.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathEnterReplacementTrustCorporationAddress, page.CanGoBack,
			EnterTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathChooseReplacementAttorneysSummary, page.CanGoBack,
			ChooseReplacementAttorneysSummary(tmpls.Get("choose_replacement_attorneys_summary.gohtml"), service, actoruid.New))
		handleWithDonor(donor.PathRemoveReplacementAttorney, page.CanGoBack,
			RemoveAttorney(tmpls.Get("remove_attorney.gohtml"), service))
		handleWithDonor(donor.PathRemoveReplacementTrustCorporation, page.CanGoBack,
			RemoveTrustCorporation(tmpls.Get("remove_attorney.gohtml"), service))
		handleWithDonor(donor.PathHowShouldReplacementAttorneysStepIn, page.CanGoBack,
			HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), donorStore))
		handleWithDonor(donor.PathHowShouldReplacementAttorneysMakeDecisions, page.CanGoBack,
			HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), donorStore))
	}

	handleWithDonor(donor.PathWhenCanTheLpaBeUsed, page.None,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), donorStore))
	handleWithDonor(donor.PathLifeSustainingTreatment, page.None,
		LifeSustainingTreatment(tmpls.Get("life_sustaining_treatment.gohtml"), donorStore))
	handleWithDonor(donor.PathRestrictions, page.None,
		Restrictions(tmpls.Get("restrictions.gohtml"), donorStore))

	{
		service := donor.NewCertificateProviderService(donorStore, reuseStore)

		handleWithDonor(donor.PathWhatACertificateProviderDoes, page.None,
			Guidance(tmpls.Get("what_a_certificate_provider_does.gohtml")))
		handleWithDonor(donor.PathChooseYourCertificateProvider, page.None,
			Guidance(tmpls.Get("choose_your_certificate_provider.gohtml")))
		handleWithDonor(donor.PathChooseNewCertificateProvider, page.None,
			RemoveCertificateProvider(tmpls.Get("choose_new_certificate_provider.gohtml"), service))
		handleWithDonor(donor.PathChooseCertificateProvider, page.None,
			ChooseCertificateProvider(tmpls.Get("choose_certificate_provider.gohtml"), service))
		handleWithDonor(donor.PathCertificateProviderDetails, page.CanGoBack,
			CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), service))
		handleWithDonor(donor.PathHowWouldCertificateProviderPreferToCarryOutTheirRole, page.CanGoBack,
			HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), donorStore, reuseStore))
		handleWithDonor(donor.PathCertificateProviderAddress, page.CanGoBack,
			CertificateProviderAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore, reuseStore))
		handleWithDonor(donor.PathHowDoYouKnowYourCertificateProvider, page.CanGoBack,
			HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), donorStore, reuseStore))
		handleWithDonor(donor.PathHowLongHaveYouKnownCertificateProvider, page.CanGoBack,
			HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), donorStore, reuseStore))
		handleWithDonor(donor.PathRemoveCertificateProvider, page.None,
			RemoveCertificateProvider(tmpls.Get("remove_certificate_provider.gohtml"), service))
		handleWithDonor(donor.PathCertificateProviderSummary, page.None,
			Guidance(tmpls.Get("certificate_provider_summary.gohtml")))
	}

	{
		service := donor.NewPeopleToNotifyService(donorStore, reuseStore)

		handleWithDonor(donor.PathDoYouWantToNotifyPeople, page.CanGoBack,
			DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), service))
		handleWithDonor(donor.PathChoosePeopleToNotify, page.CanGoBack,
			ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), service))
		handleWithDonor(donor.PathEnterPersonToNotify, page.CanGoBack,
			EnterPersonToNotify(tmpls.Get("enter_person_to_notify.gohtml"), service))
		handleWithDonor(donor.PathEnterPersonToNotifyAddress, page.CanGoBack,
			EnterPersonToNotifyAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathChoosePeopleToNotifySummary, page.CanGoBack,
			ChoosePeopleToNotifySummary(tmpls.Get("choose_people_to_notify_summary.gohtml"), service))
		handleWithDonor(donor.PathRemovePersonToNotify, page.CanGoBack,
			RemovePersonToNotify(tmpls.Get("remove_person_to_notify.gohtml"), service))
	}

	{
		service := donor.NewCorrespondentService(donorStore, reuseStore, eventClient)

		handleWithDonor(donor.PathAddCorrespondent, page.None,
			AddCorrespondent(tmpls.Get("add_correspondent.gohtml"), service))
		handleWithDonor(donor.PathChooseCorrespondent, page.CanGoBack,
			ChooseCorrespondent(tmpls.Get("choose_correspondent.gohtml"), service))
		handleWithDonor(donor.PathEnterCorrespondentDetails, page.CanGoBack,
			EnterCorrespondentDetails(tmpls.Get("enter_correspondent_details.gohtml"), service))
		handleWithDonor(donor.PathEnterCorrespondentAddress, page.CanGoBack,
			EnterCorrespondentAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, service))
		handleWithDonor(donor.PathRemoveCorrespondent, page.CanGoBack,
			RemoveCorrespondent(tmpls.Get("remove_correspondent.gohtml"), service))
		handleWithDonor(donor.PathCorrespondentSummary, page.None,
			Guidance(tmpls.Get("correspondent_summary.gohtml")))
	}

	handleWithDonor(donor.PathGettingHelpSigning, page.CanGoBack,
		Guidance(tmpls.Get("getting_help_signing.gohtml")))
	handleWithDonor(donor.PathYourAuthorisedSignatory, page.CanGoBack,
		YourAuthorisedSignatory(tmpls.Get("your_authorised_signatory.gohtml"), donorStore, actoruid.New))
	handleWithDonor(donor.PathYourIndependentWitness, page.CanGoBack,
		YourIndependentWitness(tmpls.Get("your_independent_witness.gohtml"), donorStore, actoruid.New))
	handleWithDonor(donor.PathYourIndependentWitnessMobile, page.CanGoBack,
		YourIndependentWitnessMobile(tmpls.Get("your_independent_witness_mobile.gohtml"), donorStore))
	handleWithDonor(donor.PathYourIndependentWitnessAddress, page.CanGoBack,
		YourIndependentWitnessAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))

	handleWithDonor(donor.PathYouMustBeOver18ToComplete, page.CanGoBack,
		YouMustBeOver18ToComplete(tmpls.Get("you_must_be_over_18_to_complete.gohtml")))
	handleWithDonor(donor.PathYouCannotSignYourLpaYet, page.CanGoBack,
		YouCannotSignYourLpaYet(tmpls.Get("you_cannot_sign_your_lpa_yet.gohtml")))
	handleWithDonor(donor.PathConfirmYourCertificateProviderIsNotRelated, page.CanGoBack,
		ConfirmYourCertificateProviderIsNotRelated(tmpls.Get("confirm_your_certificate_provider_is_not_related.gohtml"), donorStore, time.Now))
	handleWithDonor(donor.PathCheckYourLpa, page.CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), donorStore, shareCodeSender, notifyClient, certificateProviderStore, scheduledStore, time.Now, certificateProviderStartURL))
	handleWithDonor(donor.PathLpaDetailsSaved, page.CanGoBack,
		LpaDetailsSaved(tmpls.Get("lpa_details_saved.gohtml")))

	handleWithDonor(donor.PathAboutPayment, page.None,
		Guidance(tmpls.Get("about_payment.gohtml")))
	handleWithDonor(donor.PathAreYouApplyingForFeeDiscountOrExemption, page.CanGoBack,
		AreYouApplyingForFeeDiscountOrExemption(tmpls.Get("are_you_applying_for_a_different_fee_type.gohtml"), payer, donorStore))
	handleWithDonor(donor.PathWhichFeeTypeAreYouApplyingFor, page.CanGoBack,
		WhichFeeTypeAreYouApplyingFor(tmpls.Get("which_fee_type_are_you_applying_for.gohtml"), donorStore))
	handleWithDonor(donor.PathPreviousApplicationNumber, page.CanGoBack,
		PreviousApplicationNumber(tmpls.Get("previous_application_number.gohtml"), donorStore))
	handleWithDonor(donor.PathPreviousFee, page.CanGoBack,
		PreviousFee(tmpls.Get("previous_fee.gohtml"), payer, donorStore))
	handleWithDonor(donor.PathCostOfRepeatApplication, page.CanGoBack,
		CostOfRepeatApplication(tmpls.Get("cost_of_repeat_application.gohtml"), donorStore))
	handleWithDonor(donor.PathEvidenceRequired, page.CanGoBack,
		Guidance(tmpls.Get("evidence_required.gohtml")))
	handleWithDonor(donor.PathHowWouldYouLikeToSendEvidence, page.CanGoBack,
		HowWouldYouLikeToSendEvidence(tmpls.Get("how_would_you_like_to_send_evidence.gohtml"), donorStore))
	handleWithDonor(donor.PathUploadEvidence, page.CanGoBack,
		UploadEvidence(tmpls.Get("upload_evidence.gohtml"), logger, payer, documentStore))
	handleWithDonor(donor.PathSendUsYourEvidenceByPost, page.CanGoBack,
		SendUsYourEvidenceByPost(tmpls.Get("send_us_your_evidence_by_post.gohtml"), payer, eventClient))
	handleWithDonor(donor.PathPayFee, page.None,
		payer)
	handleWithDonor(donor.PathPaymentConfirmation, page.None,
		PaymentConfirmation(logger, payClient, donorStore, sessionStore, shareCodeSender, eventClient, notifyClient))
	handleWithDonor(donor.PathPaymentSuccessful, page.None,
		Guidance(tmpls.Get("payment_successful.gohtml")))
	handleWithDonor(donor.PathEvidenceSuccessfullyUploaded, page.None,
		Guidance(tmpls.Get("evidence_successfully_uploaded.gohtml")))
	handleWithDonor(donor.PathWhatHappensNextRepeatApplicationNoFee, page.None,
		Guidance(tmpls.Get("what_happens_next_repeat_application_no_fee.gohtml")))
	handleWithDonor(donor.PathPendingPayment, page.None,
		Guidance(tmpls.Get("pending_payment.gohtml")))

	handleWithDonor(donor.PathConfirmYourIdentity, page.CanGoBack,
		ConfirmYourIdentity(tmpls.Get("confirm_your_identity.gohtml"), donorStore))
	handleWithDonor(donor.PathHowWillYouConfirmYourIdentity, page.None,
		HowWillYouConfirmYourIdentity(tmpls.Get("how_will_you_confirm_your_identity.gohtml"), donorStore, eventClient))
	handleWithDonor(donor.PathCompletingYourIdentityConfirmation, page.None,
		CompletingYourIdentityConfirmation(tmpls.Get("completing_your_identity_confirmation.gohtml")))
	handleWithDonor(donor.PathIdentityWithOneLogin, page.CanGoBack,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.String))
	handleWithDonor(donor.PathIdentityWithOneLoginCallback, page.CanGoBack,
		IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, scheduledStore, eventClient))
	handleWithDonor(donor.PathIdentityDetails, page.None,
		IdentityDetails(tmpls.Get("identity_details.gohtml"), donorStore, eventClient))
	handleWithDonor(donor.PathIdentityDetailsUpdated, page.None,
		Guidance(tmpls.Get("identity_details_updated.gohtml")))
	handleWithDonor(donor.PathRegisterWithCourtOfProtection, page.None,
		RegisterWithCourtOfProtection(tmpls.Get("register_with_court_of_protection.gohtml"), donorStore, eventClient))

	handleWithDonor(donor.PathUnableToConfirmIdentity, page.None,
		Guidance(tmpls.Get("unable_to_confirm_identity.gohtml")))
	handleWithDonor(donor.PathChooseSomeoneToVouchForYou, page.CanGoBack,
		ChooseSomeoneToVouchForYou(tmpls.Get("choose_someone_to_vouch_for_you.gohtml"), donorStore))
	handleWithDonor(donor.PathEnterVoucher, page.CanGoBack,
		EnterVoucher(tmpls.Get("enter_voucher.gohtml"), donorStore, actoruid.New))
	handleWithDonor(donor.PathConfirmPersonAllowedToVouch, page.CanGoBack,
		ConfirmPersonAllowedToVouch(tmpls.Get("confirm_person_allowed_to_vouch.gohtml"), donorStore))
	handleWithDonor(donor.PathCheckYourDetails, page.CanGoBack,
		CheckYourDetails(tmpls.Get("check_your_details.gohtml"), shareCodeSender, donorStore))
	handleWithDonor(donor.PathWeHaveContactedVoucher, page.None,
		Guidance(tmpls.Get("we_have_contacted_voucher.gohtml")))
	handleWithDonor(donor.PathWhatYouCanDoNow, page.CanGoBack,
		WhatYouCanDoNow(tmpls.Get("what_you_can_do_now.gohtml"), donorStore, voucherStore))
	handleWithDonor(donor.PathWhatYouCanDoNowExpired, page.CanGoBack,
		WhatYouCanDoNowExpired(tmpls.Get("what_you_can_do_now_expired.gohtml"), donorStore))
	handleWithDonor(donor.PathWhatHappensNextRegisteringWithCourtOfProtection, page.None,
		Guidance(tmpls.Get("what_happens_next_registering_with_court_of_protection.gohtml")))
	handleWithDonor(donor.PathAreYouSureYouNoLongerNeedVoucher, page.CanGoBack,
		AreYouSureYouNoLongerNeedVoucher(tmpls.Get("are_you_sure_you_no_longer_need_voucher.gohtml"), donorStore, notifyClient))
	handleWithDonor(donor.PathWeHaveInformedVoucherNoLongerNeeded, page.None,
		Guidance(tmpls.Get("we_have_informed_voucher_no_longer_needed.gohtml")))
	handleWithDonor(donor.PathResendVoucherAccessCode, page.CanGoBack,
		ResendVoucherAccessCode(tmpls.Get("resend_witness_code.gohtml"), shareCodeSender))
	handleWithDonor(donor.PathChangeDonorMobileNumber, page.CanGoBack,
		YourMobile(tmpls.Get("your_mobile.gohtml"), donorStore, shareCodeSender))
	handleWithDonor(donor.PathChangeDonorEmail, page.CanGoBack,
		YourEmail(tmpls.Get("your_email.gohtml"), donorStore, shareCodeSender))

	handleWithDonor(donor.PathHowToSignYourLpa, page.None,
		Guidance(tmpls.Get("how_to_sign_your_lpa.gohtml")))
	handleWithDonor(donor.PathYourLpaLanguage, page.CanGoBack,
		YourLpaLanguage(tmpls.Get("your_lpa_language.gohtml"), donorStore))
	handleWithDonor(donor.PathReadYourLpa, page.CanGoBack,
		ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), bundle))
	handleWithDonor(donor.PathLpaYourLegalRightsAndResponsibilities, page.CanGoBack,
		Guidance(tmpls.Get("your_legal_rights_and_responsibilities.gohtml")))
	handleWithDonor(donor.PathSignYourLpa, page.CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), donorStore, scheduledStore, time.Now))
	handleWithDonor(donor.PathSignTheLpaOnBehalf, page.CanGoBack,
		SignYourLpa(tmpls.Get("sign_the_lpa_on_behalf.gohtml"), donorStore, scheduledStore, time.Now))
	handleWithDonor(donor.PathWitnessingYourSignature, page.None,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), witnessCodeSender, donorStore))
	handleWithDonor(donor.PathWitnessingAsIndependentWitness, page.None,
		WitnessingAsIndependentWitness(tmpls.Get("witnessing_as_independent_witness.gohtml"), donorStore, time.Now))
	handleWithDonor(donor.PathResendIndependentWitnessCode, page.CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithDonor(donor.PathChangeIndependentWitnessMobileNumber, page.CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithDonor(donor.PathWitnessingAsCertificateProvider, page.None,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), donorStore, shareCodeSender, lpaStoreClient, eventClient, time.Now))
	handleWithDonor(donor.PathResendCertificateProviderCode, page.CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithDonor(donor.PathChangeCertificateProviderMobileNumber, page.CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithDonor(donor.PathYouHaveSubmittedYourLpa, page.None,
		Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml")))

	handleWithDonor(donor.PathWarningInterruption, page.CanGoBack,
		WarningInterruption(tmpls.Get("warning_interruption.gohtml")))

	handleWithDonor(donor.PathProgress, page.None,
		Progress(tmpls.Get("progress.gohtml"), lpaStoreResolvingService, progressTracker, certificateProviderStore, voucherStore, donorStore, time.Now))

	handleWithDonor(donor.PathUploadEvidenceSSE, page.None,
		UploadEvidenceSSE(documentStore, logger, 3*time.Minute, 2*time.Second, time.Now))
}

func makeHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, donorStartURL string) func(page.Path, page.HandleOpt, page.Handler) {
	return func(path page.Path, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.Page = path.Format()
			appData.ActorType = actor.TypeDonor

			if opt&page.RequireSession != 0 {
				loginSession, err := store.Login(r)
				if err != nil {
					http.Redirect(w, r, donorStartURL, http.StatusFound)
					return
				}

				appData.SessionID = loginSession.SessionID()
				appData.LoginSessionEmail = loginSession.Email
				appData.HasLpas = loginSession.HasLPAs

				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID, Email: appData.LoginSessionEmail})
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeLpaHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, donorStore DonorStore, donorStartURL string) func(donor.Path, page.HandleOpt, Handler) {
	return func(path donor.Path, opt page.HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			loginSession, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, donorStartURL, http.StatusFound)
				return
			}

			appData := appcontext.DataFromContext(ctx)
			appData.CanGoBack = opt&page.CanGoBack != 0
			appData.ActorType = actor.TypeDonor
			appData.LpaID = r.PathValue("id")
			appData.SessionID = loginSession.SessionID()
			appData.LoginSessionEmail = loginSession.Email
			appData.HasLpas = loginSession.HasLPAs

			sessionData, err := appcontext.SessionFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = appcontext.ContextWithSession(ctx, sessionData)
			} else {
				sessionData = &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID}
				ctx = appcontext.ContextWithSession(ctx, sessionData)
			}

			if loginSession.OrganisationID != "" {
				sessionData.OrganisationID = loginSession.OrganisationID
				sessionData.Email = loginSession.Email
			}

			appData.Page = path.Format(appData.LpaID)

			lpa, err := donorStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !path.CanGoTo(lpa) {
				if donor.PathTaskList.CanGoTo(lpa) {
					donor.PathTaskList.Redirect(w, r, appData, lpa)
				} else {
					page.PathDashboard.Redirect(w, r, appData)
				}
				return
			}

			if lpa.Donor.Email == "" && loginSession.OrganisationID == "" {
				lpa.Donor.Email = loginSession.Email
				err = donorStore.Put(ctx, lpa)

				if err != nil {
					errorHandler(w, r, err)
					return
				}
			}

			if loginSession.OrganisationID != "" {
				appData.SupporterData = &appcontext.SupporterData{
					LpaType:          lpa.Type,
					DonorFullName:    lpa.Donor.FullName(),
					OrganisationName: loginSession.OrganisationName,
				}
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData)), lpa); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
