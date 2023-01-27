package page

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
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

func (l Lang) String() string {
	if l == Cy {
		return welshAbbreviation
	}

	return englishAbbreviation
}

func CacheControlHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000")
		h.ServeHTTP(w, r)
	})
}

func IsLpaPath(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	return path != Paths.Dashboard
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
	GetAll(context.Context, string, interface{}) error
	Get(context.Context, string, string, interface{}) error
	Put(context.Context, string, string, interface{}) error
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
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

func postFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

type AppData struct {
	Page             string
	Query            string
	Localizer        localize.Localizer
	Lang             Lang
	CookieConsentSet bool
	CanGoBack        bool
	SessionID        string
	RumConfig        RumConfig
	StaticHash       string
	Paths            AppPaths
	LpaID            string
}

func (d AppData) Redirect(w http.ResponseWriter, r *http.Request, lpa *Lpa, url string) error {
	if lpa != nil && d.LpaID == "" {
		d.LpaID = lpa.ID
	}

	// as a shortcut for when you don't have an Lpa but know the transition is fine we allow passing nil
	if lpa == nil || lpa.CanGoTo(url) {
		http.Redirect(w, r, d.BuildUrl(url), http.StatusFound)
	} else {
		http.Redirect(w, r, d.BuildUrl(Paths.TaskList), http.StatusFound)
	}

	return nil
}

func (d AppData) BuildUrl(url string) string {
	if d.Lang == Cy {
		return "/" + welshAbbreviation + d.BuildUrlWithoutLang(url)
	}

	return d.BuildUrlWithoutLang(url)
}

func (d AppData) BuildUrlWithoutLang(url string) string {
	if IsLpaPath(url) {
		return "/lpa/" + d.LpaID + url
	}

	return url
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
) http.Handler {
	rootMux := http.NewServeMux()

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}

	handleRoot := makeHandle(rootMux, logger, sessionStore, localizer, lang, rumConfig, staticHash, paths, None)

	rootMux.Handle(paths.TestingStart, testingStart(sessionStore, lpaStore, random.String))
	rootMux.Handle(paths.Root, Root(paths))

	handleRoot(paths.Start, None,
		Guidance(tmpls.Get("start.gohtml"), paths.Auth, nil))

	handleRoot(paths.Dashboard, RequireSession,
		Dashboard(tmpls.Get("dashboard.gohtml"), lpaStore))

	lpaMux := http.NewServeMux()

	rootMux.Handle("/lpa/", routeToLpa(lpaMux))

	handleLpa := makeHandle(lpaMux, logger, sessionStore, localizer, lang, rumConfig, staticHash, paths, RequireSession)

	handleLpa(paths.YourDetails, None,
		YourDetails(tmpls.Get("your_details.gohtml"), lpaStore, sessionStore))
	handleLpa(paths.YourAddress, None,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))
	handleLpa(paths.HowWouldYouLikeToBeContacted, None,
		HowWouldYouLikeToBeContacted(tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), lpaStore))
	handleLpa(paths.LpaType, None,
		LpaType(tmpls.Get("lpa_type.gohtml"), lpaStore))
	handleLpa(paths.WhoIsTheLpaFor, None,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), lpaStore))

	handleLpa(paths.TaskList, None,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStore))

	handleLpa(paths.ChooseAttorneys, CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), lpaStore, random.String))
	handleLpa(paths.ChooseAttorneysAddress, CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, lpaStore))
	handleLpa(paths.ChooseAttorneysSummary, CanGoBack,
		ChooseAttorneysSummary(logger, tmpls.Get("choose_attorneys_summary.gohtml"), lpaStore))
	handleLpa(paths.RemoveAttorney, CanGoBack,
		RemoveAttorney(logger, tmpls.Get("remove_attorney.gohtml"), lpaStore))
	handleLpa(paths.HowShouldAttorneysMakeDecisions, CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), lpaStore))

	handleLpa(paths.DoYouWantReplacementAttorneys, CanGoBack,
		WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), lpaStore))
	handleLpa(paths.ChooseReplacementAttorneys, CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), lpaStore, random.String))
	handleLpa(paths.ChooseReplacementAttorneysAddress, CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_replacement_attorneys_address.gohtml"), addressClient, lpaStore))
	handleLpa(paths.ChooseReplacementAttorneysSummary, CanGoBack,
		ChooseReplacementAttorneysSummary(logger, tmpls.Get("choose_replacement_attorneys_summary.gohtml"), lpaStore))
	handleLpa(paths.RemoveReplacementAttorney, CanGoBack,
		RemoveReplacementAttorney(logger, tmpls.Get("remove_replacement_attorney.gohtml"), lpaStore))
	handleLpa(paths.HowShouldReplacementAttorneysStepIn, CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), lpaStore))
	handleLpa(paths.HowShouldReplacementAttorneysMakeDecisions, CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), lpaStore))

	handleLpa(paths.WhenCanTheLpaBeUsed, CanGoBack,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), lpaStore))
	handleLpa(paths.Restrictions, CanGoBack,
		Restrictions(tmpls.Get("restrictions.gohtml"), lpaStore))
	handleLpa(paths.WhoDoYouWantToBeCertificateProviderGuidance, CanGoBack,
		WhoDoYouWantToBeCertificateProviderGuidance(tmpls.Get("who_do_you_want_to_be_certificate_provider_guidance.gohtml"), lpaStore))
	handleLpa(paths.CertificateProviderDetails, CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), lpaStore))
	handleLpa(paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, CanGoBack,
		HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), lpaStore))
	handleLpa(paths.CertificateProviderAddress, CanGoBack,
		CertificateProviderAddress(logger, tmpls.Get("certificate_provider_address.gohtml"), addressClient, lpaStore))
	handleLpa(paths.HowDoYouKnowYourCertificateProvider, CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), lpaStore))
	handleLpa(paths.HowLongHaveYouKnownCertificateProvider, CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), lpaStore))

	handleLpa(paths.DoYouWantToNotifyPeople, CanGoBack,
		DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), lpaStore))
	handleLpa(paths.ChoosePeopleToNotify, CanGoBack,
		ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), lpaStore, random.String))
	handleLpa(paths.ChoosePeopleToNotifyAddress, CanGoBack,
		ChoosePeopleToNotifyAddress(logger, tmpls.Get("choose_people_to_notify_address.gohtml"), addressClient, lpaStore))
	handleLpa(paths.ChoosePeopleToNotifySummary, CanGoBack,
		ChoosePeopleToNotifySummary(logger, tmpls.Get("choose_people_to_notify_summary.gohtml"), lpaStore))
	handleLpa(paths.RemovePersonToNotify, CanGoBack,
		RemovePersonToNotify(logger, tmpls.Get("remove_person_to_notify.gohtml"), lpaStore))

	handleLpa(paths.CheckYourLpa, CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), lpaStore))

	handleLpa(paths.AboutPayment, CanGoBack,
		AboutPayment(logger, tmpls.Get("about_payment.gohtml"), sessionStore, payClient, appPublicUrl, random.String, lpaStore))
	handleLpa(paths.PaymentConfirmation, CanGoBack,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, lpaStore, sessionStore))

	handleLpa(paths.HowToConfirmYourIdentityAndSign, CanGoBack,
		Guidance(tmpls.Get("how_to_confirm_your_identity_and_sign.gohtml"), Paths.WhatYoullNeedToConfirmYourIdentity, lpaStore))
	handleLpa(paths.WhatYoullNeedToConfirmYourIdentity, CanGoBack,
		Guidance(tmpls.Get("what_youll_need_to_confirm_your_identity.gohtml"), Paths.SelectYourIdentityOptions, lpaStore))

	for path, page := range map[string]int{
		paths.SelectYourIdentityOptions:  0,
		paths.SelectYourIdentityOptions1: 1,
		paths.SelectYourIdentityOptions2: 2,
	} {
		handleLpa(path, CanGoBack,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), lpaStore, page))
	}

	handleLpa(paths.YourChosenIdentityOptions, CanGoBack,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), lpaStore))
	handleLpa(paths.IdentityWithYoti, CanGoBack,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), lpaStore, yotiClient, yotiScenarioID))
	handleLpa(paths.IdentityWithYotiCallback, CanGoBack,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, lpaStore))
	handleLpa(paths.IdentityWithOneLogin, CanGoBack,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleLpa(paths.IdentityWithOneLoginCallback, CanGoBack,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))

	for path, identityOption := range map[string]IdentityOption{
		paths.IdentityWithPassport:                 Passport,
		paths.IdentityWithBiometricResidencePermit: BiometricResidencePermit,
		paths.IdentityWithDrivingLicencePaper:      DrivingLicencePaper,
		paths.IdentityWithDrivingLicencePhotocard:  DrivingLicencePhotocard,
		paths.IdentityWithOnlineBankAccount:        OnlineBankAccount,
	} {
		handleLpa(path, CanGoBack,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), identityOption))
	}

	handleLpa(paths.ReadYourLpa, CanGoBack,
		ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), lpaStore))
	handleLpa(paths.SignYourLpa, CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), lpaStore))
	handleLpa(paths.WitnessingYourSignature, CanGoBack,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), lpaStore, notifyClient, random.Code, time.Now))
	handleLpa(paths.WitnessingAsCertificateProvider, CanGoBack,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), lpaStore, time.Now))
	handleLpa(paths.YouHaveSubmittedYourLpa, CanGoBack,
		Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml"), paths.TaskList, lpaStore))

	handleLpa(paths.Progress, CanGoBack,
		Guidance(tmpls.Get("lpa_progress.gohtml"), paths.Dashboard, lpaStore))

	return rootMux
}

func testingStart(store sessions.Store, lpaStore LpaStore, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub := randomString(12)
		sessionID := base64.StdEncoding.EncodeToString([]byte(sub))

		session, _ := store.Get(r, "session")
		session.Values = map[interface{}]interface{}{"sub": sub, "email": "simulate-delivered@notifications.service.gov.uk"}
		_ = store.Save(r, w, session)

		ctx := contextWithSessionData(r.Context(), &sessionData{SessionID: sessionID})

		lpa, _ := lpaStore.Create(ctx)

		if r.FormValue("withDonorDetails") != "" || r.FormValue("completeLpa") != "" {
			lpa.You = MakePerson()
			lpa.WhoFor = "me"
			lpa.Type = "pfa"
			lpa.Tasks.YourDetails = TaskCompleted
		}

		if r.FormValue("withAttorney") != "" {
			lpa.Attorneys = []Attorney{MakeAttorney("John")}

			lpa.Tasks.ChooseAttorneys = TaskCompleted
		}

		if r.FormValue("withAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa.Attorneys = []Attorney{
				MakeAttorney("John"),
				MakeAttorney("Joan"),
			}

			lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			lpa.Tasks.ChooseAttorneys = TaskCompleted
		}

		if r.FormValue("withIncompleteAttorneys") != "" {
			withAddress := MakeAttorney("John")
			withAddress.ID = "with-address"
			withoutAddress := MakeAttorney("Joan")
			withoutAddress.ID = "without-address"
			withoutAddress.Address = place.Address{}

			lpa.Attorneys = []Attorney{
				withAddress,
				withoutAddress,
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

		if r.FormValue("howAttorneysAct") != "" {
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

		if r.FormValue("withReplacementAttorneys") != "" || r.FormValue("completeLpa") != "" {
			lpa.ReplacementAttorneys = []Attorney{
				MakeAttorney("Jane"),
				MakeAttorney("Jorge"),
			}
			lpa.WantReplacementAttorneys = "yes"
			lpa.HowReplacementAttorneysMakeDecisions = JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct
			lpa.Tasks.ChooseReplacementAttorneys = TaskCompleted
		}

		if r.FormValue("whenCanBeUsedComplete") != "" || r.FormValue("completeLpa") != "" {
			lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered
			lpa.Tasks.WhenCanTheLpaBeUsed = TaskCompleted
		}

		if r.FormValue("withRestrictions") != "" || r.FormValue("completeLpa") != "" {
			lpa.Restrictions = "Some restrictions on how Attorneys act"
			lpa.Tasks.Restrictions = TaskCompleted
		}

		if r.FormValue("withCP") == "1" || r.FormValue("completeLpa") != "" {
			lpa.CertificateProvider = MakeCertificateProvider("Barbara")
			lpa.Tasks.CertificateProvider = TaskCompleted
		}

		if r.FormValue("withPeopleToNotify") == "1" || r.FormValue("completeLpa") != "" {
			lpa.PeopleToNotify = []PersonToNotify{
				MakePersonToNotify("Joanna"),
				MakePersonToNotify("Jonathan"),
			}
			lpa.DoYouWantToNotifyPeople = "yes"
			lpa.Tasks.PeopleToNotify = TaskCompleted
		}

		if r.FormValue("lpaChecked") == "1" || r.FormValue("completeLpa") != "" {
			lpa.Checked = true
			lpa.HappyToShare = true
			lpa.Tasks.CheckYourLpa = TaskCompleted
		}

		if r.FormValue("paymentComplete") == "1" {
			paySession, _ := store.Get(r, PayCookieName)
			paySession.Values = map[interface{}]interface{}{PayCookiePaymentIdValueKey: random.String(12)}
			_ = store.Save(r, w, paySession)
			lpa.Tasks.PayForLpa = TaskCompleted
		}

		if r.FormValue("idConfirmedAndSigned") == "1" || r.FormValue("completeLpa") != "" {
			lpa.OneLoginUserData = identity.UserData{
				OK:          true,
				RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				FullName:    "Jose Smith",
			}

			lpa.WantToApplyForLpa = true
			lpa.CPWitnessedDonorSign = true
			lpa.Submitted = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
			lpa.CPWitnessCodeValidated = true
			lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

		}

		if r.FormValue("withPayment") == "1" || r.FormValue("completeLpa") != "" {
			lpa.Tasks.PayForLpa = TaskCompleted
		}

		_ = lpaStore.Put(ctx, lpa)

		if r.FormValue("cookiesAccepted") == "1" {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		random.UseTestCode = true

		AppData{}.Redirect(w, r.WithContext(ctx), lpa, r.FormValue("redirect"))
	}
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, logger Logger, store sessions.Store, localizer localize.Localizer, lang Lang, rumConfig RumConfig, staticHash string, paths AppPaths, defaultOptions handleOpt) func(string, handleOpt, Handler) {
	return func(path string, opt handleOpt, h Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := AppData{
				Page:       path,
				Query:      queryString(r),
				Localizer:  localizer,
				Lang:       lang,
				CanGoBack:  opt&CanGoBack != 0,
				RumConfig:  rumConfig,
				StaticHash: staticHash,
				Paths:      paths,
			}

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

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(sub))

				data := sessionDataFromContext(ctx)
				if data != nil {
					data.SessionID = appData.SessionID
					ctx = contextWithSessionData(ctx, data)

					appData.LpaID = data.LpaID
				} else {
					ctx = contextWithSessionData(ctx, &sessionData{SessionID: appData.SessionID})
				}
			}

			_, cookieErr := r.Cookie("cookies-consent")
			appData.CookieConsentSet = cookieErr != http.ErrNoCookie
			appData.Localizer.ShowTranslationKeys = r.FormValue("showTranslationKeys") == "1"

			if err := h(appData, w, r.WithContext(ctx)); err != nil {
				str := fmt.Sprintf("Error rendering page for path '%s': %s", path, err.Error())

				logger.Print(str)
				http.Error(w, "Encountered an error", http.StatusInternalServerError)
			}
		})
	}
}

func routeToLpa(mux http.Handler) http.HandlerFunc {
	const prefixLength = len("/lpa/")

	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 4)
		if len(parts) != 4 {
			http.NotFound(w, r)
			return
		}

		id, path := parts[2], "/"+parts[3]

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		if len(r.URL.RawPath) > prefixLength+len(id) {
			r2.URL.RawPath = r.URL.RawPath[prefixLength+len(id):]
		}

		mux.ServeHTTP(w, r2.WithContext(contextWithSessionData(r2.Context(), &sessionData{LpaID: id})))
	}
}

func queryString(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return fmt.Sprintf("?%s", r.URL.RawQuery)
	} else {
		return ""
	}
}
