package page

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

type RumConfig struct {
	GuestRoleArn      string
	Endpoint          string
	ApplicationRegion string
	IdentityPoolID    string
	ApplicationID     string
}

type Lang int

func CacheControlHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000")
		h.ServeHTTP(w, r)
	})
}

func (l Lang) Redirect(w http.ResponseWriter, r *http.Request, lpa *Lpa, url string) error {
	// as a shortcut for when you don't have an Lpa but know the transition is fine we allow passing nil
	if lpa == nil || lpa.CanGoTo(url) {
		http.Redirect(w, r, l.BuildUrl(url), http.StatusFound)
	} else {
		http.Redirect(w, r, l.BuildUrl(Paths.TaskList), http.StatusFound)
	}
	return nil
}

func (l Lang) String() string {
	if l == Cy {
		return welshAbbreviation
	}

	return englishAbbreviation
}

func (l Lang) BuildUrl(url string) string {
	if l == Cy {
		return "/" + welshAbbreviation + url
	} else {
		return url
	}
}

const (
	En Lang = iota
	Cy
	englishAbbreviation = "en"
	welshAbbreviation   = "cy"
)

type Logger interface {
	Print(v ...interface{})
}

type DataStore interface {
	Get(context.Context, string, interface{}) error
	Put(context.Context, string, interface{}) error
}

type YotiClient interface {
	IsTest() bool
	SdkID() string
	User(string) (identity.UserData, error)
}

type PayClient interface {
	CreatePayment(body pay.CreatePaymentBody) (pay.CreatePaymentResponse, error)
	GetPayment(paymentId string) (pay.GetPaymentResponse, error)
}

type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

type OneLoginClient interface {
	loginClient
	authRedirectClient
}

func postFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

type AppData struct {
	Page                string
	Query               string
	Localizer           localize.Localizer
	Lang                Lang
	CookieConsentSet    bool
	CanGoBack           bool
	SessionID           string
	RumConfig           RumConfig
	StaticHash          string
	Paths               AppPaths
	IsProduction        bool
	ShowTranslationKeys bool
}

type Handler func(data AppData, w http.ResponseWriter, r *http.Request) error

func App(
	logger Logger,
	localizer localize.Localizer,
	lang Lang,
	tmpls template.Templates,
	sessionStore sessions.Store,
	dataStore DataStore,
	appPublicUrl string,
	payClient PayClient,
	yotiClient YotiClient,
	yotiScenarioID string,
	notifyClient NotifyClient,
	addressClient AddressClient,
	rumConfig RumConfig,
	staticHash string,
	paths AppPaths,
	oneLoginClient OneLoginClient,
	isProduction bool,
) http.Handler {
	mux := http.NewServeMux()

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}

	handle := makeHandle(mux, logger, sessionStore, localizer, lang, rumConfig, staticHash, paths, isProduction)

	mux.Handle(paths.TestingStart, testingStart(sessionStore, lpaStore))
	mux.Handle(paths.Root, Root(paths))

	handle(paths.Start, None,
		Guidance(tmpls.Get("start.gohtml"), paths.Auth, nil))

	handle(paths.Dashboard, RequireSession,
		Dashboard(tmpls.Get("dashboard.gohtml"), lpaStore))
	handle(paths.LpaType, RequireSession,
		LpaType(tmpls.Get("lpa_type.gohtml"), lpaStore))
	handle(paths.WhoIsTheLpaFor, RequireSession,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), lpaStore))

	handle(paths.YourDetails, RequireSession,
		YourDetails(tmpls.Get("your_details.gohtml"), lpaStore, sessionStore))
	handle(paths.YourAddress, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))
	handle(paths.HowWouldYouLikeToBeContacted, RequireSession,
		HowWouldYouLikeToBeContacted(tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), lpaStore))

	handle(paths.TaskList, RequireSession,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStore))

	handle(paths.ChooseAttorneys, RequireSession|CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), lpaStore, random.String))
	handle(paths.ChooseAttorneysAddress, RequireSession|CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, lpaStore))
	handle(paths.ChooseAttorneysSummary, RequireSession|CanGoBack,
		ChooseAttorneysSummary(logger, tmpls.Get("choose_attorneys_summary.gohtml"), lpaStore))
	handle(paths.RemoveAttorney, RequireSession|CanGoBack,
		RemoveAttorney(logger, tmpls.Get("remove_attorney.gohtml"), lpaStore))
	handle(paths.HowShouldAttorneysMakeDecisions, RequireSession|CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), lpaStore))

	handle(paths.DoYouWantReplacementAttorneys, RequireSession|CanGoBack,
		WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), lpaStore))
	handle(paths.ChooseReplacementAttorneys, RequireSession|CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), lpaStore, random.String))
	handle(paths.ChooseReplacementAttorneysAddress, RequireSession|CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_replacement_attorneys_address.gohtml"), addressClient, lpaStore))
	handle(paths.ChooseReplacementAttorneysSummary, RequireSession|CanGoBack,
		ChooseReplacementAttorneysSummary(logger, tmpls.Get("choose_replacement_attorneys_summary.gohtml"), lpaStore))
	handle(paths.RemoveReplacementAttorney, RequireSession|CanGoBack,
		RemoveReplacementAttorney(logger, tmpls.Get("remove_replacement_attorney.gohtml"), lpaStore))
	handle(paths.HowShouldReplacementAttorneysStepIn, RequireSession|CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), lpaStore))
	handle(paths.HowShouldReplacementAttorneysMakeDecisions, RequireSession|CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), lpaStore))

	handle(paths.WhenCanTheLpaBeUsed, RequireSession|CanGoBack,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), lpaStore))
	handle(paths.Restrictions, RequireSession|CanGoBack,
		Restrictions(tmpls.Get("restrictions.gohtml"), lpaStore))
	handle(paths.WhoDoYouWantToBeCertificateProviderGuidance, RequireSession|CanGoBack,
		WhoDoYouWantToBeCertificateProviderGuidance(tmpls.Get("who_do_you_want_to_be_certificate_provider_guidance.gohtml"), lpaStore))
	handle(paths.CertificateProviderDetails, RequireSession|CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), lpaStore))
	handle(paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, RequireSession|CanGoBack,
		HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), lpaStore))
	handle(paths.CertificateProviderAddress, RequireSession|CanGoBack,
		CertificateProviderAddress(logger, tmpls.Get("certificate_provider_address.gohtml"), addressClient, lpaStore))
	handle(paths.HowDoYouKnowYourCertificateProvider, RequireSession|CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), lpaStore))
	handle(paths.HowLongHaveYouKnownCertificateProvider, RequireSession|CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), lpaStore))

	handle(paths.DoYouWantToNotifyPeople, RequireSession|CanGoBack,
		DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), lpaStore))
	handle(paths.ChoosePeopleToNotify, RequireSession|CanGoBack,
		ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), lpaStore, random.String))
	handle(paths.ChoosePeopleToNotifyAddress, RequireSession|CanGoBack,
		ChoosePeopleToNotifyAddress(logger, tmpls.Get("choose_people_to_notify_address.gohtml"), addressClient, lpaStore))
	handle(paths.ChoosePeopleToNotifySummary, RequireSession|CanGoBack,
		ChoosePeopleToNotifySummary(logger, tmpls.Get("choose_people_to_notify_summary.gohtml"), lpaStore))
	handle(paths.RemovePersonToNotify, RequireSession|CanGoBack,
		RemovePersonToNotify(logger, tmpls.Get("remove_person_to_notify.gohtml"), lpaStore))

	handle(paths.CheckYourLpa, RequireSession|CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), lpaStore))

	handle(paths.AboutPayment, RequireSession|CanGoBack,
		AboutPayment(logger, tmpls.Get("about_payment.gohtml"), sessionStore, payClient, appPublicUrl, random.String, lpaStore))
	handle(paths.PaymentConfirmation, RequireSession|CanGoBack,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, lpaStore, sessionStore))

	handle(paths.HowToConfirmYourIdentityAndSign, RequireSession|CanGoBack,
		Guidance(tmpls.Get("how_to_confirm_your_identity_and_sign.gohtml"), Paths.WhatYoullNeedToConfirmYourIdentity, lpaStore))
	handle(paths.WhatYoullNeedToConfirmYourIdentity, RequireSession|CanGoBack,
		Guidance(tmpls.Get("what_youll_need_to_confirm_your_identity.gohtml"), Paths.SelectYourIdentityOptions, lpaStore))

	for path, page := range map[string]int{
		paths.SelectYourIdentityOptions:  0,
		paths.SelectYourIdentityOptions1: 1,
		paths.SelectYourIdentityOptions2: 2,
	} {
		handle(path, RequireSession|CanGoBack,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), lpaStore, page))
	}

	handle(paths.YourChosenIdentityOptions, RequireSession|CanGoBack,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), lpaStore))
	handle(paths.IdentityWithYoti, RequireSession|CanGoBack,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), lpaStore, yotiClient, yotiScenarioID))
	handle(paths.IdentityWithYotiCallback, RequireSession|CanGoBack,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, lpaStore))
	handle(paths.IdentityWithOneLogin, RequireSession|CanGoBack,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handle(paths.IdentityWithOneLoginCallback, RequireSession|CanGoBack,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))

	for path, identityOption := range map[string]IdentityOption{
		paths.IdentityWithPassport:                 Passport,
		paths.IdentityWithBiometricResidencePermit: BiometricResidencePermit,
		paths.IdentityWithDrivingLicencePaper:      DrivingLicencePaper,
		paths.IdentityWithDrivingLicencePhotocard:  DrivingLicencePhotocard,
		paths.IdentityWithOnlineBankAccount:        OnlineBankAccount,
	} {
		handle(path, RequireSession|CanGoBack,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), identityOption))
	}

	handle(paths.ReadYourLpa, RequireSession|CanGoBack,
		ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), lpaStore))
	handle(paths.SignYourLpa, RequireSession|CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), lpaStore))
	handle(paths.WitnessingYourSignature, RequireSession|CanGoBack,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), lpaStore, notifyClient, random.Code, time.Now))
	handle(paths.WitnessingAsCertificateProvider, RequireSession|CanGoBack,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), lpaStore, time.Now))
	handle(paths.YouHaveSubmittedYourLpa, RequireSession|CanGoBack,
		Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml"), paths.TaskList, lpaStore))

	return mux
}

func testingStart(store sessions.Store, lpaStore LpaStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Values = map[interface{}]interface{}{"sub": random.String(12), "email": "simulate-delivered@notifications.service.gov.uk"}
		_ = store.Save(r, w, session)

		sessionID := base64.StdEncoding.EncodeToString([]byte(session.Values["sub"].(string)))
		var lpa *Lpa

		if r.FormValue("paymentComplete") == "1" {
			paySession, _ := store.Get(r, PayCookieName)
			paySession.Values = map[interface{}]interface{}{PayCookiePaymentIdValueKey: random.String(12)}
			_ = store.Save(r, w, paySession)
		}

		if r.FormValue("withPayment") == "1" {
			if lpa == nil {
				lpa, _ = lpaStore.Get(r.Context(), sessionID)
			}

			lpa.Tasks.PayForLpa = TaskCompleted
		}

		if r.FormValue("withAttorney") == "1" {
			if lpa == nil {
				lpa, _ = lpaStore.Get(r.Context(), sessionID)
			}

			lpa.Attorneys = []Attorney{
				{
					ID:          "with-address",
					FirstNames:  "John",
					LastName:    "Smith",
					Email:       "aa@example.org",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
					Address: place.Address{
						Line1:      "2 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
				},
			}

			lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			lpa.Tasks.ChooseAttorneys = TaskCompleted
		}

		if r.FormValue("withIncompleteAttorneys") == "1" {
			if lpa == nil {
				lpa, _ = lpaStore.Get(r.Context(), sessionID)
			}

			lpa.Attorneys = []Attorney{
				{
					ID:          "with-address",
					FirstNames:  "John",
					LastName:    "Smith",
					Email:       "aa@example.org",
					DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
					Address: place.Address{
						Line1:      "2 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
				},
				{
					ID:          "without-address",
					FirstNames:  "Joan",
					LastName:    "Smith",
					Email:       "bb@example.org",
					DateOfBirth: time.Date(1998, time.January, 2, 3, 4, 5, 6, time.UTC),
					Address:     place.Address{},
				},
			}

			lpa.ReplacementAttorneys = lpa.Attorneys
			lpa.Type = LpaTypePropertyFinance
			lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered

			lpa.HowAttorneysMakeDecisions = JointlyAndSeverally

			lpa.WantReplacementAttorneys = "yes"
			lpa.HowReplacementAttorneysMakeDecisions = JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct

			lpa.Tasks.ChooseAttorneys = TaskInProgress
			lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
		}

		if r.FormValue("withCP") == "1" {
			if lpa == nil {
				lpa, _ = lpaStore.Get(r.Context(), sessionID)
			}

			lpa.CertificateProvider = CertificateProvider{
				FirstNames:              "Barbara",
				LastName:                "Smith",
				Email:                   "b@example.org",
				Mobile:                  "07535111111",
				DateOfBirth:             time.Date(1997, time.January, 2, 3, 4, 5, 6, time.UTC),
				Relationship:            "friend",
				RelationshipDescription: "",
				RelationshipLength:      "gte-2-years",
			}
			lpa.Tasks.CertificateProvider = TaskCompleted
		}

		if r.FormValue("howAttorneysAct") != "" {
			if lpa == nil {
				lpa, _ = lpaStore.Get(r.Context(), sessionID)
			}

			switch r.FormValue("howAttorneysAct") {
			case Jointly:
				lpa.HowAttorneysMakeDecisions = Jointly
			case JointlyAndSeverally:
				lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			default:
				lpa.HowAttorneysMakeDecisions = JointlyForSomeSeverallyForOthers
				lpa.HowAttorneysMakeDecisionsDetails = "some details"
			}
		}

		if lpa != nil {
			_ = lpaStore.Put(r.Context(), sessionID, lpa)
		}

		if r.FormValue("cookiesAccepted") == "1" {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		random.UseTestCode = true

		http.Redirect(w, r, r.FormValue("redirect"), http.StatusFound)
	}
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, logger Logger, store sessions.Store, localizer localize.Localizer, lang Lang, rumConfig RumConfig, staticHash string, paths AppPaths, isProduction bool) func(string, handleOpt, Handler) {
	return func(path string, opt handleOpt, h Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			sessionID := ""

			if opt&RequireSession != 0 {
				session, err := store.Get(r, "session")
				if err != nil {
					logger.Print(err)
					http.Redirect(w, r, paths.Start, http.StatusFound)
					return
				}

				sub, ok := session.Values["sub"].(string)
				if !ok {
					logger.Print("sub missing from session")
					http.Redirect(w, r, paths.Start, http.StatusFound)
					return
				}

				sessionID = base64.StdEncoding.EncodeToString([]byte(sub))
			}

			_, cookieErr := r.Cookie("cookies-consent")

			if r.FormValue("showTranslationKeys") == "1" && !isProduction {
				localizer.showTranslationKeys = true
			} else {
				localizer.showTranslationKeys = false
			}

			if err := h(AppData{
				Page:                path,
				Query:               queryString(r),
				Localizer:           localizer,
				Lang:                lang,
				SessionID:           sessionID,
				CookieConsentSet:    cookieErr != http.ErrNoCookie,
				CanGoBack:           opt&CanGoBack != 0,
				RumConfig:           rumConfig,
				StaticHash:          staticHash,
				Paths:               paths,
				IsProduction:        isProduction,
				ShowTranslationKeys: r.Form.Get("showTranslationKeys") == "1",
			}, w, r); err != nil {
				str := fmt.Sprintf("Error rendering page for path '%s': %s", path, err.Error())

				logger.Print(str)
				http.Error(w, "Encountered an error", http.StatusInternalServerError)
			}
		})
	}
}

func queryString(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return fmt.Sprintf("?%s", r.URL.RawQuery)
	} else {
		return ""
	}
}
