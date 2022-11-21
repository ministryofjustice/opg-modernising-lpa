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

type Lang int

func (l Lang) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, l.BuildUrl(url), code)
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
	TemplateID(string) string
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
) http.Handler {
	mux := http.NewServeMux()

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}

	handle := makeHandle(mux, logger, sessionStore, localizer, lang)

	mux.Handle("/testing-start", testingStart(sessionStore, lpaStore))
	mux.Handle("/", Root())

	handle(startPath, None,
		Guidance(tmpls.Get("start.gohtml"), AuthPath, nil))
	handle(lpaTypePath, RequireSession,
		LpaType(tmpls.Get("lpa_type.gohtml"), lpaStore))
	handle(whoIsTheLpaForPath, RequireSession,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), lpaStore))

	handle(yourDetailsPath, RequireSession,
		YourDetails(tmpls.Get("your_details.gohtml"), lpaStore, sessionStore))
	handle(yourAddressPath, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))
	handle(howWouldYouLikeToBeContactedPath, RequireSession,
		HowWouldYouLikeToBeContacted(tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), lpaStore))

	handle(taskListPath, RequireSession,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStore))

	handle(chooseAttorneysPath, RequireSession|CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), lpaStore, random.String))
	handle(chooseAttorneysAddressPath, RequireSession|CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, lpaStore))
	handle(chooseAttorneysSummaryPath, RequireSession|CanGoBack,
		ChooseAttorneysSummary(logger, tmpls.Get("choose_attorneys_summary.gohtml"), lpaStore))
	handle(removeAttorneyPath, RequireSession|CanGoBack,
		RemoveAttorney(logger, tmpls.Get("remove_attorney.gohtml"), lpaStore))
	handle(howShouldAttorneysMakeDecisionsPath, RequireSession|CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), lpaStore))

	handle(wantReplacementAttorneysPath, RequireSession|CanGoBack,
		WantReplacementAttorneys(tmpls.Get("want_replacement_attorneys.gohtml"), lpaStore))
	handle(chooseReplacementAttorneysPath, RequireSession|CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), lpaStore, random.String))
	handle(chooseReplacementAttorneysAddressPath, RequireSession|CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_replacement_attorneys_address.gohtml"), addressClient, lpaStore))
	handle(chooseReplacementAttorneysSummaryPath, RequireSession|CanGoBack,
		ChooseReplacementAttorneysSummary(logger, tmpls.Get("choose_replacement_attorneys_summary.gohtml"), lpaStore))
	handle(removeReplacementAttorneyPath, RequireSession|CanGoBack,
		RemoveReplacementAttorney(logger, tmpls.Get("remove_replacement_attorney.gohtml"), lpaStore))
	handle(howShouldReplacementAttorneysStepInPath, RequireSession|CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), lpaStore))
	handle(howShouldReplacementAttorneysMakeDecisionsPath, RequireSession|CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), lpaStore))

	handle(whenCanTheLpaBeUsedPath, RequireSession|CanGoBack,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), lpaStore))
	handle(restrictionsPath, RequireSession|CanGoBack,
		Restrictions(tmpls.Get("restrictions.gohtml"), lpaStore))
	handle(whoDoYouWantToBeCertificateProviderGuidancePath, RequireSession|CanGoBack,
		WhoDoYouWantToBeCertificateProviderGuidance(tmpls.Get("who_do_you_want_to_be_certificate_provider_guidance.gohtml"), lpaStore))
	handle(certificateProviderDetailsPath, RequireSession|CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), lpaStore))
	handle(howDoYouKnowYourCertificateProviderPath, RequireSession|CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), lpaStore))
	handle(howLongHaveYouKnownCertificateProviderPath, RequireSession|CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), lpaStore))

	handle(aboutPaymentPath, RequireSession|CanGoBack,
		AboutPayment(logger, tmpls.Get("about_payment.gohtml"), sessionStore, payClient, appPublicUrl, random.String))
	handle(paymentConfirmationPath, RequireSession|CanGoBack,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, lpaStore, sessionStore))

	handle(checkYourLpaPath, RequireSession|CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), lpaStore))

	handle(whatHappensNextPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("what_happens_next.gohtml"), aboutPaymentPath, lpaStore))

	handle(selectYourIdentityOptionsPath, RequireSession|CanGoBack,
		SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), lpaStore))
	handle(yourChosenIdentityOptionsPath, RequireSession|CanGoBack,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), lpaStore))
	handle(identityWithYotiPath, RequireSession|CanGoBack,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), lpaStore, yotiClient, yotiScenarioID))
	handle(identityWithYotiCallbackPath, RequireSession|CanGoBack,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, lpaStore))
	handle(identityWithPassportPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, Passport))
	handle(identityWithDrivingLicencePath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, DrivingLicence))
	handle(identityWithGovernmentGatewayAccountPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, GovernmentGatewayAccount))
	handle(identityWithDwpAccountPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, DwpAccount))
	handle(identityWithOnlineBankAccountPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, OnlineBankAccount))
	handle(identityWithUtilityBillPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, UtilityBill))
	handle(identityWithCouncilTaxBillPath, RequireSession|CanGoBack,
		IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, CouncilTaxBill))
	handle(whatHappensWhenSigningPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("what_happens_when_signing.gohtml"), howToSignPath, lpaStore))
	handle(howToSignPath, RequireSession|CanGoBack,
		HowToSign(tmpls.Get("how_to_sign.gohtml"), lpaStore, notifyClient, random.Code))
	handle(readYourLpaPath, RequireSession|CanGoBack,
		ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), lpaStore))
	handle(signingConfirmationPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("signing_confirmation.gohtml"), taskListPath, lpaStore))
	handle(dashboardPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("dashboard.gohtml"), "", lpaStore))

	return mux
}

func testingStart(store sessions.Store, lpaStore LpaStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Values = map[interface{}]interface{}{"sub": random.String(12), "email": "simulate-delivered@notifications.service.gov.uk"}
		_ = store.Save(r, w, session)

		if r.FormValue("paymentComplete") == "1" {
			paySession, _ := store.Get(r, PayCookieName)
			paySession.Values = map[interface{}]interface{}{PayCookiePaymentIdValueKey: random.String(12)}
			_ = store.Save(r, w, paySession)
		}

		if r.FormValue("withAttorneys") == "1" {
			sessionID := base64.StdEncoding.EncodeToString([]byte(session.Values["sub"].(string)))
			lpa, _ := lpaStore.Get(r.Context(), sessionID)

			lpa.Attorneys = []Attorney{
				{
					ID:          "completed-address",
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
					ID:          "empty-address",
					FirstNames:  "Joan",
					LastName:    "Smith",
					Email:       "bb@example.org",
					DateOfBirth: time.Date(1998, time.January, 2, 3, 4, 5, 6, time.UTC),
					Address:     place.Address{},
				},
			}

			lpa.ReplacementAttorneys = lpa.Attorneys

			_ = lpaStore.Put(r.Context(), sessionID, lpa)
		}

		if r.FormValue("howAttorneysAct") != "" {
			sessionID := base64.StdEncoding.EncodeToString([]byte(session.Values["sub"].(string)))
			lpa, _ := lpaStore.Get(r.Context(), sessionID)

			switch r.FormValue("howAttorneysAct") {
			case Jointly:
				lpa.HowAttorneysMakeDecisions = Jointly
			case JointlyAndSeverally:
				lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
			default:
				lpa.HowAttorneysMakeDecisions = JointlyForSomeSeverallyForOthers
				lpa.HowAttorneysMakeDecisionsDetails = "some details"
			}

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

func makeHandle(mux *http.ServeMux, logger Logger, store sessions.Store, localizer localize.Localizer, lang Lang) func(string, handleOpt, Handler) {
	return func(path string, opt handleOpt, h Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			sessionID := ""

			if opt&RequireSession != 0 {
				session, err := store.Get(r, "session")
				if err != nil {
					logger.Print(err)
					http.Redirect(w, r, startPath, http.StatusFound)
					return
				}

				sub, ok := session.Values["sub"].(string)
				if !ok {
					logger.Print("sub missing from session")
					http.Redirect(w, r, startPath, http.StatusFound)
					return
				}

				sessionID = base64.StdEncoding.EncodeToString([]byte(sub))
			}

			_, cookieErr := r.Cookie("cookies-consent")

			if err := h(AppData{
				Page:             path,
				Query:            queryString(r),
				Localizer:        localizer,
				Lang:             lang,
				SessionID:        sessionID,
				CookieConsentSet: cookieErr != http.ErrNoCookie,
				CanGoBack:        opt&CanGoBack != 0,
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
