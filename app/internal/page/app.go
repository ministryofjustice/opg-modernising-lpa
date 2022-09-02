package page

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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
	requireSession := makeRequireSession(logger, sessionStore)

	mux.Handle("/testing-start", testingStart(sessionStore))

	mux.Handle("/", Root())
	mux.Handle(startPath,
		Start(logger, localizer, lang, tmpls.Get("start.gohtml")))
	mux.Handle(lpaTypePath, requireSession(
		LpaType(logger, localizer, lang, tmpls.Get("lpa_type.gohtml"), dataStore)))
	mux.Handle(whoIsTheLpaForPath, requireSession(
		WhoIsTheLpaFor(logger, localizer, lang, tmpls.Get("who_is_the_lpa_for.gohtml"), dataStore)))
	mux.Handle(donorDetailsPath, requireSession(
		DonorDetails(logger, localizer, lang, tmpls.Get("donor_details.gohtml"), dataStore)))
	mux.Handle(donorAddressPath, requireSession(
		DonorAddress(logger, localizer, lang, tmpls.Get("donor_address.gohtml"), addressClient, dataStore)))
	mux.Handle(howWouldYouLikeToBeContactedPath, requireSession(
		HowWouldYouLikeToBeContacted(logger, localizer, lang, tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), dataStore)))
	mux.Handle(taskListPath, requireSession(
		TaskList(logger, localizer, lang, tmpls.Get("task_list.gohtml"), dataStore)))

	return mux
}

func testingStart(store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "params")
		session.Values = map[interface{}]interface{}{"email": "testing@example.com"}
		_ = store.Save(r, w, session)

		http.Redirect(w, r, r.FormValue("redirect"), http.StatusFound)
	}
}

func makeRequireSession(logger Logger, store sessions.Store) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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

			sessionID := base64.StdEncoding.EncodeToString([]byte(email))

			r = r.WithContext(context.WithValue(r.Context(), sessionKey{}, sessionID))
			h.ServeHTTP(w, r)
		}
	}
}

type sessionKey struct{}

func cookieConsentSet(r *http.Request) bool {
	_, err := r.Cookie("cookies-consent")

	return err != http.ErrNoCookie
}

func sessionID(r *http.Request) string {
	value := r.Context().Value(sessionKey{})

	return value.(string)
}
