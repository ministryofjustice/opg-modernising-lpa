package donor

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpastore.Lpa, error)
}

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error

type Template func(io.Writer, interface{}) error

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type DonorStore interface {
	Get(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Latest(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Put(ctx context.Context, donor *actor.DonorProvidedDetails) error
	Delete(ctx context.Context) error
	Link(ctx context.Context, data actor.ShareCodeData, donorEmail string) error
}

type GetDonorStore interface {
	Get(context.Context) (*actor.DonorProvidedDetails, error)
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
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
	SendCertificateProviderInvite(context.Context, page.AppData, page.CertificateProviderInvite) error
	SendCertificateProviderPrompt(context.Context, page.AppData, *actor.DonorProvidedDetails) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

type NotifyClient interface {
	SendActorSMS(ctx context.Context, to, lpaUID string, sms notify.SMS) error
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
	SendToCertificateProvider(context.Context, *actor.DonorProvidedDetails, page.Localizer) error
	SendToIndependentWitness(context.Context, *actor.DonorProvidedDetails, page.Localizer) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

type RequestSigner interface {
	Sign(context.Context, *http.Request, string) error
}

type Localizer interface {
	page.Localizer
}

type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document) error
	Delete(context.Context, page.Document) error
	DeleteInfectedDocuments(context.Context, page.Documents) error
	Create(context.Context, *actor.DonorProvidedDetails, string, []byte) (page.Document, error)
	Submit(context.Context, *actor.DonorProvidedDetails, page.Documents) error
}

type EventClient interface {
	SendReducedFeeRequested(ctx context.Context, e event.ReducedFeeRequested) error
	SendPaymentReceived(ctx context.Context, e event.PaymentReceived) error
	SendUidRequested(ctx context.Context, e event.UidRequested) error
	SendPreviousApplicationLinked(ctx context.Context, e event.PreviousApplicationLinked) error
}

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	SendLpa(ctx context.Context, details *actor.DonorProvidedDetails) error
	Lpa(ctx context.Context, lpaUID string) (*lpastore.Lpa, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, code string) (actor.ShareCodeData, error)
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type ProgressTracker interface {
	Progress(lpa *lpastore.Lpa) page.Progress
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
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
) {
	payer := Pay(logger, sessionStore, donorStore, payClient, random.String, appPublicURL)

	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Login, page.None,
		page.Login(oneLoginClient, sessionStore, random.String, page.Paths.LoginCallback))
	handleRoot(page.Paths.LoginCallback, page.None,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.Dashboard, dashboardStore, actor.TypeDonor))
	handleRoot(page.Paths.EnterAccessCode, page.RequireSession,
		EnterAccessCode(tmpls.Get("enter_access_code.gohtml"), shareCodeStore, donorStore))

	handleWithDonor := makeLpaHandle(rootMux, sessionStore, errorHandler, donorStore)

	handleWithDonor(page.Paths.DeleteThisLpa, page.None,
		DeleteLpa(tmpls.Get("delete_this_lpa.gohtml"), donorStore))
	handleWithDonor(page.Paths.WithdrawThisLpa, page.None,
		WithdrawLpa(tmpls.Get("withdraw_this_lpa.gohtml"), donorStore, time.Now))

	handleWithDonor(page.Paths.MakeANewLPA, page.None,
		Guidance(tmpls.Get("make_a_new_lpa.gohtml")))
	handleWithDonor(page.Paths.YourDetails, page.None,
		YourDetails(tmpls.Get("your_details.gohtml"), donorStore, sessionStore))
	handleWithDonor(page.Paths.YourName, page.None,
		YourName(tmpls.Get("your_name.gohtml"), donorStore))
	handleWithDonor(page.Paths.YourDateOfBirth, page.None,
		YourDateOfBirth(tmpls.Get("your_date_of_birth.gohtml"), donorStore))
	handleWithDonor(page.Paths.YourAddress, page.None,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.WeHaveUpdatedYourDetails, page.None,
		Guidance(tmpls.Get("we_have_updated_your_details.gohtml")))
	handleWithDonor(page.Paths.YourPreferredLanguage, page.None,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), donorStore))
	handleWithDonor(page.Paths.LpaType, page.None,
		LpaType(tmpls.Get("lpa_type.gohtml"), donorStore, eventClient))
	handleWithDonor(page.Paths.CheckYouCanSign, page.None,
		CheckYouCanSign(tmpls.Get("check_you_can_sign.gohtml"), donorStore))
	handleWithDonor(page.Paths.NeedHelpSigningConfirmation, page.None,
		Guidance(tmpls.Get("need_help_signing_confirmation.gohtml")))

	handleWithDonor(page.Paths.TaskList, page.None,
		TaskList(tmpls.Get("task_list.gohtml"), evidenceReceivedStore))

	handleWithDonor(page.Paths.ChooseAttorneysGuidance, page.None,
		Guidance(tmpls.Get("choose_attorneys_guidance.gohtml")))
	handleWithDonor(page.Paths.ChooseAttorneys, page.CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), donorStore, actoruid.New))
	handleWithDonor(page.Paths.ChooseAttorneysAddress, page.CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.EnterTrustCorporation, page.CanGoBack,
		EnterTrustCorporation(tmpls.Get("enter_trust_corporation.gohtml"), donorStore))
	handleWithDonor(page.Paths.EnterTrustCorporationAddress, page.CanGoBack,
		EnterTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.ChooseAttorneysSummary, page.CanGoBack,
		ChooseAttorneysSummary(tmpls.Get("choose_attorneys_summary.gohtml")))
	handleWithDonor(page.Paths.RemoveAttorney, page.CanGoBack,
		RemoveAttorney(tmpls.Get("remove_attorney.gohtml"), donorStore))
	handleWithDonor(page.Paths.RemoveTrustCorporation, page.CanGoBack,
		RemoveTrustCorporation(tmpls.Get("remove_attorney.gohtml"), donorStore, false))
	handleWithDonor(page.Paths.HowShouldAttorneysMakeDecisions, page.CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), donorStore))

	handleWithDonor(page.Paths.DoYouWantReplacementAttorneys, page.None,
		WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), donorStore))
	handleWithDonor(page.Paths.ChooseReplacementAttorneys, page.CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), donorStore, actoruid.New))
	handleWithDonor(page.Paths.ChooseReplacementAttorneysAddress, page.CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.EnterReplacementTrustCorporation, page.CanGoBack,
		EnterReplacementTrustCorporation(tmpls.Get("enter_replacement_trust_corporation.gohtml"), donorStore))
	handleWithDonor(page.Paths.EnterReplacementTrustCorporationAddress, page.CanGoBack,
		EnterReplacementTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.ChooseReplacementAttorneysSummary, page.CanGoBack,
		ChooseReplacementAttorneysSummary(tmpls.Get("choose_replacement_attorneys_summary.gohtml")))
	handleWithDonor(page.Paths.RemoveReplacementAttorney, page.CanGoBack,
		RemoveReplacementAttorney(tmpls.Get("remove_attorney.gohtml"), donorStore))
	handleWithDonor(page.Paths.RemoveReplacementTrustCorporation, page.CanGoBack,
		RemoveTrustCorporation(tmpls.Get("remove_attorney.gohtml"), donorStore, true))
	handleWithDonor(page.Paths.HowShouldReplacementAttorneysStepIn, page.CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), donorStore))
	handleWithDonor(page.Paths.HowShouldReplacementAttorneysMakeDecisions, page.CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), donorStore))

	handleWithDonor(page.Paths.WhenCanTheLpaBeUsed, page.None,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), donorStore))
	handleWithDonor(page.Paths.LifeSustainingTreatment, page.None,
		LifeSustainingTreatment(tmpls.Get("life_sustaining_treatment.gohtml"), donorStore))
	handleWithDonor(page.Paths.Restrictions, page.None,
		Restrictions(tmpls.Get("restrictions.gohtml"), donorStore))

	handleWithDonor(page.Paths.WhatACertificateProviderDoes, page.None,
		Guidance(tmpls.Get("what_a_certificate_provider_does.gohtml")))
	handleWithDonor(page.Paths.ChooseYourCertificateProvider, page.None,
		Guidance(tmpls.Get("choose_your_certificate_provider.gohtml")))
	handleWithDonor(page.Paths.ChooseNewCertificateProvider, page.None,
		ChooseNewCertificateProvider(tmpls.Get("choose_new_certificate_provider.gohtml"), donorStore))
	handleWithDonor(page.Paths.CertificateProviderDetails, page.CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), donorStore, actoruid.New))
	handleWithDonor(page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, page.CanGoBack,
		HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), donorStore))
	handleWithDonor(page.Paths.CertificateProviderAddress, page.CanGoBack,
		CertificateProviderAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.HowDoYouKnowYourCertificateProvider, page.CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), donorStore))
	handleWithDonor(page.Paths.HowLongHaveYouKnownCertificateProvider, page.CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), donorStore))

	handleWithDonor(page.Paths.DoYouWantToNotifyPeople, page.CanGoBack,
		DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), donorStore))
	handleWithDonor(page.Paths.ChoosePeopleToNotify, page.CanGoBack,
		ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), donorStore, actoruid.New))
	handleWithDonor(page.Paths.ChoosePeopleToNotifyAddress, page.CanGoBack,
		ChoosePeopleToNotifyAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.ChoosePeopleToNotifySummary, page.CanGoBack,
		ChoosePeopleToNotifySummary(tmpls.Get("choose_people_to_notify_summary.gohtml")))
	handleWithDonor(page.Paths.RemovePersonToNotify, page.CanGoBack,
		RemovePersonToNotify(tmpls.Get("remove_person_to_notify.gohtml"), donorStore))

	handleWithDonor(page.Paths.AddCorrespondent, page.None,
		AddCorrespondent(tmpls.Get("add_correspondent.gohtml"), donorStore))
	handleWithDonor(page.Paths.EnterCorrespondentDetails, page.CanGoBack,
		EnterCorrespondentDetails(tmpls.Get("enter_correspondent_details.gohtml"), donorStore))
	handleWithDonor(page.Paths.EnterCorrespondentAddress, page.CanGoBack,
		EnterCorrespondentAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithDonor(page.Paths.WhoCanCorrespondentsDetailsBeSharedWith, page.CanGoBack,
		WhoCanCorrespondentsDetailsBeSharedWith(tmpls.Get("who_can_correspondents_details_be_shared_with.gohtml"), donorStore))

	handleWithDonor(page.Paths.GettingHelpSigning, page.CanGoBack,
		Guidance(tmpls.Get("getting_help_signing.gohtml")))
	handleWithDonor(page.Paths.YourAuthorisedSignatory, page.CanGoBack,
		YourAuthorisedSignatory(tmpls.Get("your_authorised_signatory.gohtml"), donorStore))
	handleWithDonor(page.Paths.YourIndependentWitness, page.CanGoBack,
		YourIndependentWitness(tmpls.Get("your_independent_witness.gohtml"), donorStore))
	handleWithDonor(page.Paths.YourIndependentWitnessMobile, page.CanGoBack,
		YourIndependentWitnessMobile(tmpls.Get("your_independent_witness_mobile.gohtml"), donorStore))
	handleWithDonor(page.Paths.YourIndependentWitnessAddress, page.CanGoBack,
		YourIndependentWitnessAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))

	handleWithDonor(page.Paths.YouCannotSignYourLpaYet, page.CanGoBack,
		YouCannotSignYourLpaYet(tmpls.Get("you_cannot_sign_your_lpa_yet.gohtml")))
	handleWithDonor(page.Paths.ConfirmYourCertificateProviderIsNotRelated, page.CanGoBack,
		ConfirmYourCertificateProviderIsNotRelated(tmpls.Get("confirm_your_certificate_provider_is_not_related.gohtml"), donorStore, time.Now))
	handleWithDonor(page.Paths.CheckYourLpa, page.CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), donorStore, shareCodeSender, notifyClient, certificateProviderStore, time.Now, appPublicURL))
	handleWithDonor(page.Paths.LpaDetailsSaved, page.CanGoBack,
		LpaDetailsSaved(tmpls.Get("lpa_details_saved.gohtml")))

	handleWithDonor(page.Paths.AboutPayment, page.None,
		Guidance(tmpls.Get("about_payment.gohtml")))
	handleWithDonor(page.Paths.AreYouApplyingForFeeDiscountOrExemption, page.CanGoBack,
		AreYouApplyingForFeeDiscountOrExemption(tmpls.Get("are_you_applying_for_a_different_fee_type.gohtml"), payer, donorStore))
	handleWithDonor(page.Paths.WhichFeeTypeAreYouApplyingFor, page.CanGoBack,
		WhichFeeTypeAreYouApplyingFor(tmpls.Get("which_fee_type_are_you_applying_for.gohtml"), donorStore))
	handleWithDonor(page.Paths.PreviousApplicationNumber, page.None,
		PreviousApplicationNumber(tmpls.Get("previous_application_number.gohtml"), donorStore, eventClient))
	handleWithDonor(page.Paths.PreviousFee, page.CanGoBack,
		PreviousFee(tmpls.Get("previous_fee.gohtml"), payer, donorStore))
	handleWithDonor(page.Paths.EvidenceRequired, page.CanGoBack,
		Guidance(tmpls.Get("evidence_required.gohtml")))
	handleWithDonor(page.Paths.HowWouldYouLikeToSendEvidence, page.CanGoBack,
		HowWouldYouLikeToSendEvidence(tmpls.Get("how_would_you_like_to_send_evidence.gohtml"), donorStore))
	handleWithDonor(page.Paths.UploadEvidence, page.CanGoBack,
		UploadEvidence(tmpls.Get("upload_evidence.gohtml"), logger, payer, documentStore))
	handleWithDonor(page.Paths.SendUsYourEvidenceByPost, page.CanGoBack,
		SendUsYourEvidenceByPost(tmpls.Get("send_us_your_evidence_by_post.gohtml"), payer, eventClient))
	handleWithDonor(page.Paths.FeeApproved, page.None,
		payer)
	handleWithDonor(page.Paths.FeeDenied, page.None,
		FeeDenied(tmpls.Get("fee_denied.gohtml"), payer))
	handleWithDonor(page.Paths.PaymentConfirmation, page.None,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, donorStore, sessionStore, shareCodeSender, lpaStoreClient, eventClient))
	handleWithDonor(page.Paths.EvidenceSuccessfullyUploaded, page.None,
		Guidance(tmpls.Get("evidence_successfully_uploaded.gohtml")))
	handleWithDonor(page.Paths.WhatHappensNextPostEvidence, page.None,
		Guidance(tmpls.Get("what_happens_next_post_evidence.gohtml")))

	handleWithDonor(page.Paths.HowToConfirmYourIdentityAndSign, page.None,
		Guidance(tmpls.Get("how_to_confirm_your_identity_and_sign.gohtml")))
	handleWithDonor(page.Paths.ProveYourIdentity, page.CanGoBack,
		Guidance(tmpls.Get("prove_your_identity.gohtml")))
	handleWithDonor(page.Paths.IdentityWithOneLogin, page.CanGoBack,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.String))
	handleWithDonor(page.Paths.IdentityWithOneLoginCallback, page.CanGoBack,
		IdentityWithOneLoginCallback(commonTmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, donorStore))

	handleWithDonor(page.Paths.ReadYourLpa, page.None,
		Guidance(tmpls.Get("read_your_lpa.gohtml")))
	handleWithDonor(page.Paths.LpaYourLegalRightsAndResponsibilities, page.CanGoBack,
		Guidance(tmpls.Get("your_legal_rights_and_responsibilities.gohtml")))
	handleWithDonor(page.Paths.SignYourLpa, page.CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), donorStore, time.Now))
	handleWithDonor(page.Paths.SignTheLpaOnBehalf, page.CanGoBack,
		SignYourLpa(tmpls.Get("sign_the_lpa_on_behalf.gohtml"), donorStore, time.Now))
	handleWithDonor(page.Paths.WitnessingYourSignature, page.None,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), witnessCodeSender, donorStore))
	handleWithDonor(page.Paths.WitnessingAsIndependentWitness, page.None,
		WitnessingAsIndependentWitness(tmpls.Get("witnessing_as_independent_witness.gohtml"), donorStore, time.Now))
	handleWithDonor(page.Paths.ResendIndependentWitnessCode, page.CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithDonor(page.Paths.ChangeIndependentWitnessMobileNumber, page.CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithDonor(page.Paths.WitnessingAsCertificateProvider, page.None,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), donorStore, shareCodeSender, lpaStoreClient, time.Now))
	handleWithDonor(page.Paths.ResendCertificateProviderCode, page.CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithDonor(page.Paths.ChangeCertificateProviderMobileNumber, page.CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithDonor(page.Paths.YouHaveSubmittedYourLpa, page.None,
		Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml")))

	handleWithDonor(page.Paths.Progress, page.CanGoBack,
		LpaProgress(tmpls.Get("lpa_progress.gohtml"), lpaStoreResolvingService, progressTracker))

	handleWithDonor(page.Paths.UploadEvidenceSSE, page.None,
		UploadEvidenceSSE(documentStore, 3*time.Minute, 2*time.Second, time.Now))
}

func makeHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler) func(page.Path, page.HandleOpt, page.Handler) {
	return func(path page.Path, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.ActorType = actor.TypeDonor

			if opt&page.RequireSession != 0 {
				session, err := store.Login(r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = session.SessionID()
				appData.LoginSessionEmail = session.Email
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID, Email: appData.LoginSessionEmail})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeLpaHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, donorStore DonorStore) func(page.LpaPath, page.HandleOpt, Handler) {
	return func(path page.LpaPath, opt page.HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			loginSession, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
				return
			}

			appData := page.AppDataFromContext(ctx)
			appData.CanGoBack = opt&page.CanGoBack != 0
			appData.ActorType = actor.TypeDonor
			appData.LpaID = r.PathValue("id")
			appData.SessionID = loginSession.SessionID()
			appData.LoginSessionEmail = loginSession.Email

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = page.ContextWithSessionData(ctx, sessionData)
			} else {
				sessionData = &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID}
				ctx = page.ContextWithSessionData(ctx, sessionData)
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

			if loginSession.OrganisationID != "" {
				appData.SupporterData = &page.SupporterData{
					LpaType:          lpa.Type,
					DonorFullName:    lpa.Donor.FullName(),
					OrganisationName: loginSession.OrganisationName,
				}
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData)), lpa); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
