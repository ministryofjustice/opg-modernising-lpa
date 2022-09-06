package page

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

type Lang int

func (l Lang) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	if l == En {
		http.Redirect(w, r, url, code)
	} else {
		http.Redirect(w, r, "/cy"+url, code)
	}
}

const (
	En Lang = iota
	Cy
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
) http.Handler {
	mux := http.NewServeMux()

	addressClient := fakeAddressClient{}
	handle := makeHandle(mux, logger, sessionStore, localizer, lang)

	mux.Handle("/testing-start", testingStart(sessionStore))
	mux.Handle("/", Root())

	handle(startPath, None,
		Start(tmpls.Get("start.gohtml")))
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

	return mux
}

func testingStart(store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Values = map[interface{}]interface{}{"sub": random.String(12)}
		_ = store.Save(r, w, session)

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
					logger.Print("email missing from session")
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
				logger.Print(err)
				http.Error(w, "an error occurred", http.StatusInternalServerError)
			}
		})
	}
}
