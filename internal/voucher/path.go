package voucher

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
)

const (
	PathTaskList = Path("/task-list")
)

type Path string

func (p Path) String() string {
	return "/voucher/{id}" + string(p)
}

func (p Path) Format(id string) string {
	return "/voucher/" + id + string(p)
}

func (p Path) Redirect(w http.ResponseWriter, r *http.Request, appData appcontext.Data, lpaID string) error {
	rurl := p.Format(lpaID)
	if fromURL := r.FormValue("from"); fromURL != "" {
		rurl = fromURL
	}

	http.Redirect(w, r, appData.Lang.URL(rurl), http.StatusFound)
	return nil
}
