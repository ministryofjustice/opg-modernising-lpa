package supporterpage

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type SearchClient interface {
	Query(ctx context.Context, req search.QueryRequest) (*search.QueryResponse, error)
	CountWithQuery(ctx context.Context, req search.CountWithQueryReq) (int, error)
}

type dashboardData struct {
	App         appcontext.Data
	Errors      validation.List
	Donors      []donordata.Provided
	CurrentPage int
	Pagination  *search.Pagination
}

func Dashboard(tmpl template.Template, donorStore DonorStore, searchClient SearchClient) Handler {
	const pageSize = 10

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
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
			App:         appData,
			Donors:      donors,
			CurrentPage: page,
			Pagination:  resp.Pagination,
		})
	}
}
