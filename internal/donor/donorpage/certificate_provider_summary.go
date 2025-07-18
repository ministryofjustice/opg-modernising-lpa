package donorpage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderSummaryData struct {
	App            appcontext.Data
	Errors         validation.List
	Donor          *donordata.Provided
	CanChangeEmail bool
}

func CertificateProviderSummary(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &certificateProviderSummaryData{
			App:   appData,
			Donor: provided,
		}

		if _, err := certificateProviderStore.OneByUID(r.Context(), provided.LpaUID); err != nil {
			if errors.Is(err, dynamo.NotFoundError{}) {
				data.CanChangeEmail = true
			} else {
				return fmt.Errorf("get certificate provider: %w", err)
			}
		}

		return tmpl(w, data)
	}
}
