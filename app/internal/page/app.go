package page

import (
	"encoding/json"
	"net/http"
	"strings"

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

func App(logger Logger, localizer localize.Localizer, lang Lang, tmpls template.Templates) http.Handler {
	mux := http.NewServeMux()

	addressClient := fakeAddressClient{}
	dataStore := fakeDataStore{logger: logger}

	mux.Handle(startPath, Start(logger, localizer, lang, tmpls.Get("start.gohtml")))
	mux.Handle(donorDetailsPath, DonorDetails(logger, localizer, lang, tmpls.Get("donor_details.gohtml"), dataStore))
	mux.Handle(donorAddressPath, DonorAddress(logger, localizer, lang, tmpls.Get("donor_address.gohtml"), addressClient, dataStore))
	mux.Handle(whoIsTheLpaForPath, WhoIsTheLpaFor(logger, localizer, lang, tmpls.Get("who_is_the_lpa_for.gohtml"), dataStore))

	return mux
}
