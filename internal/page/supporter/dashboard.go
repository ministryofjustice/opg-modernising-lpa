package supporter

import (
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

const pageSize = 10

type dashboardData struct {
	App        page.AppData
	Errors     validation.List
	Donors     []actor.DonorProvidedDetails
	Pagination *search.Pagination
}

func Dashboard(tmpl template.Template, donorStore DonorStore, searchClient *search.Client) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		page, err := strconv.Atoi(r.FormValue("page"))
		if err != nil {
			page = 1
		}

		resp, err := searchClient.Query(r.Context(), search.QueryRequest{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return err
		}

		donors, err := donorStore.GetByKeys(r.Context(), resp.Keys)
		if err != nil {
			return err
		}

		return tmpl(w, &dashboardData{
			App:        appData,
			Donors:     donors,
			Pagination: resp.Pagination,
		})
	}
}
