package page

import (
	"encoding/json"
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

type fakeAddressClient struct{}

func (c fakeAddressClient) LookupPostcode(postcode string) ([]Address, error) {
	return []Address{
		{Line1: "123 Fake Street", TownOrCity: "Someville", Postcode: postcode},
		{Line1: "456 Fake Street", TownOrCity: "Someville", Postcode: postcode},
	}, nil
}

type DataStore interface {
	Save(interface{}) error
}

type fakeDataStore struct {
	logger Logger
}

func (d fakeDataStore) Save(v interface{}) error {
	data, _ := json.Marshal(v)
	d.logger.Print(string(data))
	return nil
}

func postFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func App(logger Logger, localizer localize.Localizer, lang Lang, tmpls template.Templates, sessionStore sessions.Store) http.Handler {
	mux := http.NewServeMux()

	addressClient := fakeAddressClient{}
	dataStore := fakeDataStore{logger: logger}
	requireSession := makeRequireSession(logger, sessionStore)

	mux.HandleFunc("/testing-start", func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, "params")
		if err != nil {
			logger.Print(err)
			return
		}

		session.Values = map[interface{}]interface{}{"email": "testing@example.com"}
		if err := sessionStore.Save(r, w, session); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, r.FormValue("redirect"), http.StatusFound)
	})

	mux.Handle(startPath,
		Start(logger, localizer, lang, tmpls.Get("start.gohtml")))
	mux.Handle(whoIsTheLpaForPath, requireSession(
		WhoIsTheLpaFor(logger, localizer, lang, tmpls.Get("who_is_the_lpa_for.gohtml"), dataStore)))
	mux.Handle(donorDetailsPath, requireSession(
		DonorDetails(logger, localizer, lang, tmpls.Get("donor_details.gohtml"), dataStore)))
	mux.Handle(donorAddressPath, requireSession(
		DonorAddress(logger, localizer, lang, tmpls.Get("donor_address.gohtml"), addressClient, dataStore)))
	mux.Handle(howWouldYouLikeToBeContactedPath, requireSession(
		HowWouldYouLikeToBeContacted(logger, localizer, lang, tmpls.Get("how_would_you_like_to_be_contacted.gohtml"), dataStore)))

	return mux
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

			if _, ok := session.Values["email"].(string); !ok {
				logger.Print("email missing from session")
				http.Redirect(w, r, startPath, http.StatusFound)
				return
			}

			h.ServeHTTP(w, r)
		}
	}
}
