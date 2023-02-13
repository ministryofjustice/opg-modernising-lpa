package page

import (
	"context"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Logger interface {
	Print(v ...interface{})
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
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

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

type AppData struct {
	Page             string
	Query            string
	Localizer        localize.Localizer
	Lang             localize.Lang
	CookieConsentSet bool
	CanGoBack        bool
	SessionID        string
	RumConfig        RumConfig
	StaticHash       string
	Paths            AppPaths
	LpaID            string
	CsrfToken        string
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
	if d.Lang == localize.Cy {
		return "/" + localize.Cy.String() + d.BuildUrlWithoutLang(url)
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
