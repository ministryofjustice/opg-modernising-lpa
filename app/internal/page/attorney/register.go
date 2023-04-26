package attorney

import (
	"context"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name LpaStore --structname mockLpaStore
type LpaStore interface {
	Create(context.Context) (*page.Lpa, error)
	GetAll(context.Context) ([]*page.Lpa, error)
	Get(context.Context) (*page.Lpa, error)
	Put(context.Context, *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	GetAll(context.Context, string, interface{}) error
	Get(context.Context, string, string, interface{}) error
	Put(context.Context, string, string, interface{}) error
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	lpaStore LpaStore,
	oneLoginClient OneLoginClient,
	dataStore DataStore,
	errorHandler page.ErrorHandler,
	notifyClient NotifyClient,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Start, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml"), nil))
	handleRoot(page.Paths.Attorney.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("attorney_enter_reference_number.gohtml"), lpaStore, dataStore, sessionStore))
	handleRoot(page.Paths.Attorney.TaskList, RequireLpa,
		TaskList(tmpls.Get("attorney_task_list.gohtml"), lpaStore))
	handleRoot(page.Paths.Attorney.CheckYourName, RequireLpa,
		CheckYourName(tmpls.Get("attorney_check_your_name.gohtml"), lpaStore, notifyClient))
	handleRoot(page.Paths.Attorney.DateOfBirth, RequireLpa,
		DateOfBirth(tmpls.Get("attorney_date_of_birth.gohtml"), lpaStore))
	handleRoot(page.Paths.Attorney.Sign, RequireLpa,
		Sign(tmpls.Get("attorney_sign.gohtml"), lpaStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	RequireLpa
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ServiceName = "beAnAttorney"
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				if _, err := sesh.Attorney(store, r); err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}
			}

			if opt&RequireLpa != 0 {
				session, err := sesh.Attorney(store, r)
				if err != nil || session.DonorSessionID == "" || session.LpaID == "" || session.AttorneyID == "" {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}

				appData.SessionID = session.DonorSessionID
				appData.LpaID = session.LpaID
				appData.AttorneyID = session.AttorneyID
				appData.IsReplacementAttorney = session.IsReplacementAttorney

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{
					SessionID: appData.SessionID,
					LpaID:     appData.LpaID,
				})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func getProvidedDetails(appData page.AppData, lpa *page.Lpa) actor.AttorneyProvidedDetails {
	if appData.IsReplacementAttorney {
		if providedDetails, ok := lpa.ReplacementAttorneyProvidedDetails[appData.AttorneyID]; ok {
			return providedDetails
		}
	} else {
		if providedDetails, ok := lpa.AttorneyProvidedDetails[appData.AttorneyID]; ok {
			return providedDetails
		}
	}

	return actor.AttorneyProvidedDetails{}
}

func setProvidedDetails(appData page.AppData, lpa *page.Lpa, providedDetails actor.AttorneyProvidedDetails) {
	if appData.IsReplacementAttorney {
		if lpa.ReplacementAttorneyProvidedDetails == nil {
			lpa.ReplacementAttorneyProvidedDetails = map[string]actor.AttorneyProvidedDetails{appData.AttorneyID: providedDetails}
		} else {
			lpa.ReplacementAttorneyProvidedDetails[appData.AttorneyID] = providedDetails
		}
	} else {
		if lpa.AttorneyProvidedDetails == nil {
			lpa.AttorneyProvidedDetails = map[string]actor.AttorneyProvidedDetails{appData.AttorneyID: providedDetails}
		} else {
			lpa.AttorneyProvidedDetails[appData.AttorneyID] = providedDetails
		}
	}
}

func getTasks(appData page.AppData, lpa *page.Lpa) page.AttorneyTasks {
	if appData.IsReplacementAttorney {
		if tasks, ok := lpa.ReplacementAttorneyTasks[appData.AttorneyID]; ok {
			return tasks
		}
	} else {
		if tasks, ok := lpa.AttorneyTasks[appData.AttorneyID]; ok {
			return tasks
		}
	}

	return page.AttorneyTasks{}
}

func setTasks(appData page.AppData, lpa *page.Lpa, tasks page.AttorneyTasks) {
	if appData.IsReplacementAttorney {
		if lpa.ReplacementAttorneyTasks == nil {
			lpa.ReplacementAttorneyTasks = map[string]page.AttorneyTasks{appData.AttorneyID: tasks}
		} else {
			lpa.ReplacementAttorneyTasks[appData.AttorneyID] = tasks
		}
	} else {
		if lpa.AttorneyTasks == nil {
			lpa.AttorneyTasks = map[string]page.AttorneyTasks{appData.AttorneyID: tasks}
		} else {
			lpa.AttorneyTasks[appData.AttorneyID] = tasks
		}
	}
}
