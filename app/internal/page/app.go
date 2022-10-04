package page

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/easyid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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

type fakeAddressClient struct{}

func (c fakeAddressClient) LookupPostcode(postcode string) ([]Address, error) {
	return []Address{
		{Line1: "123 Fake Street", TownOrCity: "Someville", Postcode: postcode},
		{Line1: "456 Fake Street", TownOrCity: "Someville", Postcode: postcode},
	}, nil
}

type yotiClient interface {
	IsTest() bool
	SdkID() string
	User(string) (easyid.UserData, error)
}

func postFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

type AppData struct {
	Page             string
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
	payClient *pay.Client,
	yotiClient yotiClient,
	yotiScenarioID string,
) http.Handler {
	mux := http.NewServeMux()

	addressClient := fakeAddressClient{}

	handle := makeHandle(mux, logger, sessionStore, localizer, lang)

	mux.Handle("/testing-start", testingStart(sessionStore))
	mux.Handle("/", Root())

	handle(startPath, None,
		Guidance(tmpls.Get("start.gohtml"), AuthPath, nil))
	handle(lpaTypePath, RequireSession,
		LpaType(tmpls.Get("lpa_type.gohtml"), dataStore))
	handle(whoIsTheLpaForPath, RequireSession,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), dataStore))

	handle(yourDetailsPath, RequireSession,
		YourDetails(tmpls.Get("your_details.gohtml"), dataStore))
	handle(yourAddressPath, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, dataStore))
	handle(howWouldYouLikeToBeContactedPath, RequireSession,
		HowWouldYouLikeToBeContacted(tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), dataStore))
	handle(taskListPath, RequireSession,
		TaskList(tmpls.Get("task_list.gohtml"), dataStore))
	handle(chooseAttorneysPath, RequireSession|CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), dataStore))
	handle(chooseAttorneysAddressPath, RequireSession|CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, dataStore))
	handle(wantReplacementAttorneysPath, RequireSession|CanGoBack,
		WantReplacementAttorneys(tmpls.Get("want_replacement_attorneys.gohtml"), dataStore))
	handle(whenCanTheLpaBeUsedPath, RequireSession|CanGoBack,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), dataStore))
	handle(restrictionsPath, RequireSession|CanGoBack,
		Restrictions(tmpls.Get("restrictions.gohtml"), dataStore))
	handle(whoDoYouWantToBeCertificateProviderGuidancePath, RequireSession|CanGoBack,
		WhoDoYouWantToBeCertificateProviderGuidance(tmpls.Get("who_do_you_want_to_be_certificate_provider_guidance.gohtml"), dataStore))
	handle(certificateProviderDetailsPath, RequireSession|CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), dataStore))
	handle(howDoYouKnowYourCertificateProviderPath, RequireSession|CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), dataStore))
	handle(howLongHaveYouKnownCertificateProviderPath, RequireSession|CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), dataStore))
	handle(aboutPaymentPath, RequireSession|CanGoBack,
		AboutPayment(logger, tmpls.Get("about_payment.gohtml"), sessionStore, payClient, appPublicUrl, random.String))
	handle(checkYourLpaPath, RequireSession|CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), dataStore))

	handle(paymentConfirmationPath, RequireSession|CanGoBack,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, dataStore, sessionStore))
	handle(whatHappensNextPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("what_happens_next.gohtml"), aboutPaymentPath, dataStore))

	handle(selectYourIdentityOptionsPath, RequireSession|CanGoBack,
		SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), dataStore))
	handle(identityWithEasyIDPath, RequireSession|CanGoBack,
		IdentityWithEasyID(tmpls.Get("identity_with_easy_id.gohtml"), yotiClient, yotiScenarioID))
	handle(identityWithEasyIDCallbackPath, RequireSession|CanGoBack,
		IdentityWithEasyIDCallback(yotiClient))
	handle(whatHappensWhenSigningPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("what_happens_when_signing.gohtml"), howToSignPath, dataStore))
	handle(howToSignPath, RequireSession|CanGoBack,
		Guidance(tmpls.Get("how_to_sign.gohtml"), readYourLpaPath, dataStore))
	handle(readYourLpaPath, RequireSession|CanGoBack,
		ReadYourLpa(tmpls.Get("read_your_lpa.gohtml"), dataStore))

	return mux
}

func testingStart(store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Values = map[interface{}]interface{}{"sub": random.String(12)}
		_ = store.Save(r, w, session)

		if r.FormValue("paymentComplete") == "1" {
			paySession, _ := store.Get(r, PayCookieName)
			paySession.Values = map[interface{}]interface{}{PayCookiePaymentIdValueKey: random.String(12)}
			_ = store.Save(r, w, paySession)
		}

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
