package donor

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, details *page.Lpa) error

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	Get(context.Context) (*page.Lpa, error)
	Latest(context.Context) (*page.Lpa, error)
	Put(context.Context, *page.Lpa) error
	Delete(context.Context) error
}

type GetDonorStore interface {
	Get(context.Context) (*page.Lpa, error)
}

//go:generate mockery --testonly --inpackage --name CertificateProviderStore --structname mockCertificateProviderStore
type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
}

//go:generate mockery --testonly --inpackage --name AttorneyStore --structname mockAttorneyStore
type AttorneyStore interface {
	GetAny(ctx context.Context) ([]*actor.AttorneyProvidedDetails, error)
}

//go:generate mockery --testonly --inpackage --name EvidenceReceivedStore --structname mockEvidenceReceivedStore
type EvidenceReceivedStore interface {
	Get(context.Context) (bool, error)
}

//go:generate mockery --testonly --inpackage --name S3Client --structname mockS3Client
type S3Client interface {
	PutObject(context.Context, string, []byte) error
	DeleteObject(context.Context, string) error
}

//go:generate mockery --testonly --inpackage --name PayClient --structname mockPayClient
type PayClient interface {
	CreatePayment(body pay.CreatePaymentBody) (pay.CreatePaymentResponse, error)
	GetPayment(paymentId string) (pay.GetPaymentResponse, error)
}

//go:generate mockery --testonly --inpackage --name AddressClient --structname mockAddressClient
type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeSender --structname mockShareCodeSender
type ShareCodeSender interface {
	SendCertificateProvider(ctx context.Context, template notify.Template, appData page.AppData, identity bool, lpa *page.Lpa) error
	SendAttorneys(ctx context.Context, appData page.AppData, lpa *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.Template) string
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name WitnessCodeSender --structname mockWitnessCodeSender
type WitnessCodeSender interface {
	SendToCertificateProvider(context.Context, *page.Lpa, page.Localizer) error
	SendToIndependentWitness(context.Context, *page.Lpa, page.Localizer) error
}

//go:generate mockery --testonly --inpackage --name UidClient --structname mockUidClient
type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

//go:generate mockery --testonly --inpackage --name RequestSigner --structname mockRequestSigner
type RequestSigner interface {
	Sign(context.Context, *http.Request, string) error
}

//go:generate mockery --testonly --inpackage --name Payer --structname mockPayer
type Payer interface {
	Pay(page.AppData, http.ResponseWriter, *http.Request, *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name Localizer --structname mockLocalizer
type Localizer interface {
	Format(string, map[string]any) string
	T(string) string
	Count(messageID string, count int) string
	FormatCount(messageID string, count int, data map[string]interface{}) string
	ShowTranslationKeys() bool
	SetShowTranslationKeys(s bool)
	Possessive(s string) string
	Concat([]string, string) string
}

//go:generate mockery --testonly --inpackage --name DocumentStore --structname mockDocumentStore
type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document) error
	Delete(context.Context, page.Document) error
	DeleteInfectedDocuments(context.Context, page.Documents) error
	Create(context.Context, *page.Lpa, string, []byte) (page.Document, error)
	Submit(context.Context, *page.Lpa, page.Documents) error
}

//go:generate mockery --testonly --inpackage --name EventClient --structname mockEventClient
type EventClient interface {
	SendReducedFeeRequested(context.Context, event.ReducedFeeRequested) error
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
	notFoundHandler page.Handler,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	notifyClient NotifyClient,
	evidenceReceivedStore EvidenceReceivedStore,
	evidenceS3Client S3Client,
	documentStore DocumentStore,
	eventClient EventClient,
) {
	payer := &payHelper{
		logger:       logger,
		sessionStore: sessionStore,
		donorStore:   donorStore,
		payClient:    payClient,
		appPublicURL: appPublicURL,
		randomString: random.String,
	}

	handleRoot := makeHandle(rootMux, sessionStore, None, errorHandler)

	handleRoot(page.Paths.Login, None,
		page.Login(logger, oneLoginClient, sessionStore, random.String, page.Paths.LoginCallback))
	handleRoot(page.Paths.LoginCallback, None,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.Dashboard))

	lpaMux := http.NewServeMux()
	rootMux.Handle("/lpa/", page.RouteToPrefix("/lpa/", lpaMux, notFoundHandler))

	handleLpa := makeHandle(lpaMux, sessionStore, RequireSession, errorHandler)
	handleWithLpa := makeLpaHandle(lpaMux, sessionStore, RequireSession, errorHandler, donorStore)

	handleLpa(page.Paths.Root, None, notFoundHandler)

	handleWithLpa(page.Paths.DeleteThisLpa, None,
		DeleteLpa(tmpls.Get("delete_this_lpa.gohtml"), donorStore))
	handleWithLpa(page.Paths.WithdrawThisLpa, None,
		WithdrawLpa(tmpls.Get("withdraw_this_lpa.gohtml"), donorStore, time.Now))

	handleWithLpa(page.Paths.YourDetails, None,
		YourDetails(tmpls.Get("your_details.gohtml"), donorStore, sessionStore))
	handleWithLpa(page.Paths.YourAddress, None,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.LpaType, None,
		LpaType(tmpls.Get("lpa_type.gohtml"), donorStore))
	handleWithLpa(page.Paths.CheckYouCanSign, None,
		CheckYouCanSign(tmpls.Get("check_you_can_sign.gohtml"), donorStore))
	handleWithLpa(page.Paths.NeedHelpSigningConfirmation, None,
		Guidance(tmpls.Get("need_help_signing_confirmation.gohtml")))

	handleWithLpa(page.Paths.TaskList, None,
		TaskList(tmpls.Get("task_list.gohtml"), evidenceReceivedStore))

	handleWithLpa(page.Paths.ChooseAttorneysGuidance, None,
		Guidance(tmpls.Get("choose_attorneys_guidance.gohtml")))
	handleWithLpa(page.Paths.ChooseAttorneys, CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), donorStore, random.UuidString))
	handleWithLpa(page.Paths.ChooseAttorneysAddress, CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.EnterTrustCorporation, CanGoBack,
		EnterTrustCorporation(tmpls.Get("enter_trust_corporation.gohtml"), donorStore))
	handleWithLpa(page.Paths.EnterTrustCorporationAddress, CanGoBack,
		EnterTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.ChooseAttorneysSummary, CanGoBack,
		ChooseAttorneysSummary(tmpls.Get("choose_attorneys_summary.gohtml")))
	handleWithLpa(page.Paths.RemoveAttorney, CanGoBack,
		RemoveAttorney(logger, tmpls.Get("remove_attorney.gohtml"), donorStore))
	handleWithLpa(page.Paths.HowShouldAttorneysMakeDecisions, CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), donorStore))

	handleWithLpa(page.Paths.DoYouWantReplacementAttorneys, None,
		WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), donorStore))
	handleWithLpa(page.Paths.ChooseReplacementAttorneys, CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), donorStore, random.UuidString))
	handleWithLpa(page.Paths.ChooseReplacementAttorneysAddress, CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.EnterReplacementTrustCorporation, CanGoBack,
		EnterReplacementTrustCorporation(tmpls.Get("enter_replacement_trust_corporation.gohtml"), donorStore))
	handleWithLpa(page.Paths.EnterReplacementTrustCorporationAddress, CanGoBack,
		EnterReplacementTrustCorporationAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.ChooseReplacementAttorneysSummary, CanGoBack,
		ChooseReplacementAttorneysSummary(tmpls.Get("choose_replacement_attorneys_summary.gohtml")))
	handleWithLpa(page.Paths.RemoveReplacementAttorney, CanGoBack,
		RemoveReplacementAttorney(logger, tmpls.Get("remove_replacement_attorney.gohtml"), donorStore))
	handleWithLpa(page.Paths.HowShouldReplacementAttorneysStepIn, CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), donorStore))
	handleWithLpa(page.Paths.HowShouldReplacementAttorneysMakeDecisions, CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), donorStore))

	handleWithLpa(page.Paths.WhenCanTheLpaBeUsed, None,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), donorStore))
	handleWithLpa(page.Paths.LifeSustainingTreatment, None,
		LifeSustainingTreatment(tmpls.Get("life_sustaining_treatment.gohtml"), donorStore))
	handleWithLpa(page.Paths.Restrictions, None,
		Restrictions(tmpls.Get("restrictions.gohtml"), donorStore))

	handleWithLpa(page.Paths.WhatACertificateProviderDoes, None,
		Guidance(tmpls.Get("what_a_certificate_provider_does.gohtml")))
	handleWithLpa(page.Paths.ChooseYourCertificateProvider, None,
		Guidance(tmpls.Get("choose_your_certificate_provider.gohtml")))
	handleWithLpa(page.Paths.ChooseNewCertificateProvider, None,
		ChooseNewCertificateProvider(tmpls.Get("choose_new_certificate_provider.gohtml"), donorStore))
	handleWithLpa(page.Paths.CertificateProviderDetails, CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), donorStore))
	handleWithLpa(page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, CanGoBack,
		HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), donorStore))
	handleWithLpa(page.Paths.CertificateProviderAddress, CanGoBack,
		CertificateProviderAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.HowDoYouKnowYourCertificateProvider, CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), donorStore))
	handleWithLpa(page.Paths.HowLongHaveYouKnownCertificateProvider, CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), donorStore))

	handleWithLpa(page.Paths.DoYouWantToNotifyPeople, CanGoBack,
		DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), donorStore))
	handleWithLpa(page.Paths.ChoosePeopleToNotify, CanGoBack,
		ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), donorStore, random.UuidString))
	handleWithLpa(page.Paths.ChoosePeopleToNotifyAddress, CanGoBack,
		ChoosePeopleToNotifyAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))
	handleWithLpa(page.Paths.ChoosePeopleToNotifySummary, CanGoBack,
		ChoosePeopleToNotifySummary(tmpls.Get("choose_people_to_notify_summary.gohtml")))
	handleWithLpa(page.Paths.RemovePersonToNotify, CanGoBack,
		RemovePersonToNotify(logger, tmpls.Get("remove_person_to_notify.gohtml"), donorStore))

	handleWithLpa(page.Paths.GettingHelpSigning, CanGoBack,
		Guidance(tmpls.Get("getting_help_signing.gohtml")))
	handleWithLpa(page.Paths.YourAuthorisedSignatory, CanGoBack,
		YourAuthorisedSignatory(tmpls.Get("your_authorised_signatory.gohtml"), donorStore))
	handleWithLpa(page.Paths.YourIndependentWitness, CanGoBack,
		YourIndependentWitness(tmpls.Get("your_independent_witness.gohtml"), donorStore))
	handleWithLpa(page.Paths.YourIndependentWitnessMobile, CanGoBack,
		YourIndependentWitnessMobile(tmpls.Get("your_independent_witness_mobile.gohtml"), donorStore))
	handleWithLpa(page.Paths.YourIndependentWitnessAddress, CanGoBack,
		YourIndependentWitnessAddress(logger, tmpls.Get("choose_address.gohtml"), addressClient, donorStore))

	handleWithLpa(page.Paths.ConfirmYourCertificateProviderIsNotRelated, CanGoBack,
		ConfirmYourCertificateProviderIsNotRelated(tmpls.Get("confirm_your_certificate_provider_is_not_related.gohtml"), donorStore))
	handleWithLpa(page.Paths.CheckYourLpa, CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), donorStore, shareCodeSender, notifyClient, certificateProviderStore))
	handleWithLpa(page.Paths.LpaDetailsSaved, CanGoBack,
		LpaDetailsSaved(tmpls.Get("lpa_details_saved.gohtml")))

	handleWithLpa(page.Paths.AboutPayment, None,
		Guidance(tmpls.Get("about_payment.gohtml")))
	handleWithLpa(page.Paths.AreYouApplyingForFeeDiscountOrExemption, CanGoBack,
		AreYouApplyingForFeeDiscountOrExemption(tmpls.Get("are_you_applying_for_a_different_fee_type.gohtml"), payer, donorStore))
	handleWithLpa(page.Paths.WhichFeeTypeAreYouApplyingFor, CanGoBack,
		WhichFeeTypeAreYouApplyingFor(tmpls.Get("which_fee_type_are_you_applying_for.gohtml"), donorStore))
	handleWithLpa(page.Paths.PreviousApplicationNumber, None,
		PreviousApplicationNumber(tmpls.Get("previous_application_number.gohtml"), donorStore))
	handleWithLpa(page.Paths.PreviousFee, CanGoBack,
		PreviousFee(tmpls.Get("previous_fee.gohtml"), payer, donorStore))
	handleWithLpa(page.Paths.EvidenceRequired, CanGoBack,
		Guidance(tmpls.Get("evidence_required.gohtml")))
	handleWithLpa(page.Paths.HowWouldYouLikeToSendEvidence, CanGoBack,
		HowWouldYouLikeToSendEvidence(tmpls.Get("how_would_you_like_to_send_evidence.gohtml"), donorStore))
	handleWithLpa(page.Paths.UploadEvidence, CanGoBack,
		UploadEvidence(tmpls.Get("upload_evidence.gohtml"), payer, documentStore))
	handleWithLpa(page.Paths.SendUsYourEvidenceByPost, CanGoBack,
		SendUsYourEvidenceByPost(tmpls.Get("send_us_your_evidence_by_post.gohtml"), payer, eventClient))
	handleWithLpa(page.Paths.FeeDenied, None,
		FeeDenied(tmpls.Get("fee_denied.gohtml"), payer))
	handleWithLpa(page.Paths.PaymentConfirmation, None,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, donorStore, sessionStore))
	handleWithLpa(page.Paths.EvidenceSuccessfullyUploaded, None,
		Guidance(tmpls.Get("evidence_successfully_uploaded.gohtml")))
	handleWithLpa(page.Paths.WhatHappensNextPostEvidence, None,
		Guidance(tmpls.Get("what_happens_next_post_evidence.gohtml")))

	handleWithLpa(page.Paths.HowToConfirmYourIdentityAndSign, None,
		Guidance(tmpls.Get("how_to_confirm_your_identity_and_sign.gohtml")))
	handleWithLpa(page.Paths.ProveYourIdentity, CanGoBack,
		Guidance(tmpls.Get("prove_your_identity.gohtml")))
	handleWithLpa(page.Paths.IdentityWithOneLogin, CanGoBack,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleWithLpa(page.Paths.IdentityWithOneLoginCallback, CanGoBack,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, donorStore))

	handleWithLpa(page.Paths.ReadYourLpa, None,
		Guidance(tmpls.Get("read_your_lpa.gohtml")))
	handleWithLpa(page.Paths.LpaYourLegalRightsAndResponsibilities, CanGoBack,
		Guidance(tmpls.Get("your_legal_rights_and_responsibilities.gohtml")))
	handleWithLpa(page.Paths.SignYourLpa, CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), donorStore))
	handleWithLpa(page.Paths.SignTheLpaOnBehalf, CanGoBack,
		SignYourLpa(tmpls.Get("sign_the_lpa_on_behalf.gohtml"), donorStore))
	handleWithLpa(page.Paths.WitnessingYourSignature, None,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), witnessCodeSender, donorStore))
	handleWithLpa(page.Paths.WitnessingAsIndependentWitness, None,
		WitnessingAsIndependentWitness(tmpls.Get("witnessing_as_independent_witness.gohtml"), donorStore, time.Now))
	handleWithLpa(page.Paths.ResendIndependentWitnessCode, CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithLpa(page.Paths.ChangeIndependentWitnessMobileNumber, CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeIndependentWitness))
	handleWithLpa(page.Paths.WitnessingAsCertificateProvider, None,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), donorStore, shareCodeSender, time.Now, certificateProviderStore))
	handleWithLpa(page.Paths.ResendCertificateProviderCode, CanGoBack,
		ResendWitnessCode(tmpls.Get("resend_witness_code.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithLpa(page.Paths.ChangeCertificateProviderMobileNumber, CanGoBack,
		ChangeMobileNumber(tmpls.Get("change_mobile_number.gohtml"), witnessCodeSender, actor.TypeCertificateProvider))
	handleWithLpa(page.Paths.YouHaveSubmittedYourLpa, None,
		Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml")))

	handleWithLpa(page.Paths.Progress, CanGoBack,
		LpaProgress(tmpls.Get("lpa_progress.gohtml"), certificateProviderStore, attorneyStore))

	handleWithLpa(page.Paths.UploadEvidenceSSE, None,
		UploadEvidenceSSE(documentStore, 3*time.Minute, 2*time.Second))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, defaultOptions handleOpt, errorHandler page.ErrorHandler) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.CanGoBack = opt&CanGoBack != 0
			appData.ActorType = actor.TypeDonor

			if opt&RequireSession != 0 {
				donorSession, err := sesh.Login(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(donorSession.Sub))

				sessionData, err := page.SessionDataFromContext(ctx)

				if err == nil {
					sessionData.SessionID = appData.SessionID
					ctx = page.ContextWithSessionData(ctx, sessionData)

					appData.LpaID = sessionData.LpaID
				} else {
					ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
				}
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeLpaHandle(mux *http.ServeMux, store sesh.Store, defaultOptions handleOpt, errorHandler page.ErrorHandler, donorStore DonorStore) func(page.LpaPath, handleOpt, Handler) {
	return func(path page.LpaPath, opt handleOpt, h Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0
			appData.ActorType = actor.TypeDonor

			donorSession, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
				return
			}

			appData.SessionID = base64.StdEncoding.EncodeToString([]byte(donorSession.Sub))

			sessionData, err := page.SessionDataFromContext(ctx)

			if err == nil {
				sessionData.SessionID = appData.SessionID
				ctx = page.ContextWithSessionData(ctx, sessionData)

				appData.LpaID = sessionData.LpaID
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			appData.Page = path.Format(appData.LpaID)

			lpa, err := donorStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData)), lpa); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

type payHelper struct {
	logger       Logger
	sessionStore sessions.Store
	donorStore   DonorStore
	payClient    PayClient
	appPublicURL string
	randomString func(int) string
}

func (p *payHelper) Pay(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
	if lpa.FeeType.IsNoFee() || lpa.FeeType.IsHardshipFee() || lpa.Tasks.PayForLpa.IsMoreEvidenceRequired() {
		lpa.Tasks.PayForLpa = actor.PaymentTaskPending
		if err := p.donorStore.Put(r.Context(), lpa); err != nil {
			return err
		}

		if lpa.EvidenceDelivery.IsPost() {
			return appData.Redirect(w, r, lpa, page.Paths.WhatHappensNextPostEvidence.Format(lpa.ID))
		}

		return appData.Redirect(w, r, lpa, page.Paths.EvidenceSuccessfullyUploaded.Format(lpa.ID))
	}

	createPaymentBody := pay.CreatePaymentBody{
		Amount:      lpa.FeeAmount(),
		Reference:   p.randomString(12),
		Description: "Property and Finance LPA",
		ReturnUrl:   p.appPublicURL + appData.BuildUrl(page.Paths.PaymentConfirmation.Format(lpa.ID)),
		Email:       lpa.Donor.Email,
		Language:    appData.Lang.String(),
	}

	resp, err := p.payClient.CreatePayment(createPaymentBody)
	if err != nil {
		p.logger.Print(fmt.Sprintf("Error creating payment: %s", err.Error()))
		return err
	}

	if err = sesh.SetPayment(p.sessionStore, r, w, &sesh.PaymentSession{PaymentID: resp.PaymentId}); err != nil {
		return err
	}

	if lpa.Tasks.PayForLpa.IsDenied() {
		lpa.FeeType = pay.FullFee
		lpa.Tasks.PayForLpa = actor.PaymentTaskInProgress
		if err := p.donorStore.Put(r.Context(), lpa); err != nil {
			return err
		}
	}

	nextUrl := resp.Links["next_url"].Href
	// If URL matches expected domain for GOV UK PAY redirect there. If not,
	// redirect to the confirmation code and carry on with flow.
	if strings.HasPrefix(nextUrl, pay.PaymentPublicServiceUrl) {
		http.Redirect(w, r, nextUrl, http.StatusFound)
		return nil
	}

	return appData.Redirect(w, r, lpa, page.Paths.PaymentConfirmation.Format(lpa.ID))
}
