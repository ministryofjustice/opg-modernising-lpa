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

	handle(startPath, false,
		Start(tmpls.Get("start.gohtml")))
	handle(lpaTypePath, true,
		LpaType(tmpls.Get("lpa_type.gohtml"), dataStore))
	handle(whoIsTheLpaForPath, true,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), dataStore))
	handle(donorDetailsPath, true,
		DonorDetails(tmpls.Get("donor_details.gohtml"), dataStore))
	handle(donorAddressPath, true,
		DonorAddress(logger, tmpls.Get("donor_address.gohtml"), addressClient, dataStore))
	handle(howWouldYouLikeToBeContactedPath, true,
		HowWouldYouLikeToBeContacted(tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), dataStore))
	handle(taskListPath, true,
		TaskList(tmpls.Get("task_list.gohtml"), dataStore))
	handle(chooseAttorneysPath, true,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), dataStore))
	handle(chooseAttorneysAddressPath, true,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, dataStore))

	return mux
}

func testingStart(store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "params")
		session.Values = map[interface{}]interface{}{"email": random.String(12) + "@example.com"}
		_ = store.Save(r, w, session)

		http.Redirect(w, r, r.FormValue("redirect"), http.StatusFound)
	}
}

func makeHandle(mux *http.ServeMux, logger Logger, store sessions.Store, localizer localize.Localizer, lang Lang) func(string, bool, Handler) {
	return func(path string, requireSession bool, h Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			sessionID := ""

			if requireSession {
				session, err := store.Get(r, "params")
				if err != nil {
					logger.Print(err)
					http.Redirect(w, r, startPath, http.StatusFound)
					return
				}

				email, ok := session.Values["email"].(string)
				if !ok {
					logger.Print("email missing from session")
					http.Redirect(w, r, startPath, http.StatusFound)
					return
				}

				sessionID = base64.StdEncoding.EncodeToString([]byte(email))
			}

			_, cookieErr := r.Cookie("cookies-consent")

			if err := h(AppData{
				Page:             path,
				Localizer:        localizer,
				Lang:             lang,
				SessionID:        sessionID,
				CookieConsentSet: cookieErr != http.ErrNoCookie,
			}, w, r); err != nil {
				logger.Print(err)
				http.Error(w, "an error occurred", http.StatusInternalServerError)
			}
		})
	}
}
