package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

func DependencyHealthCheck(logger Logger, uidClient UidClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := uidClient.CreateCase(r.Context(), &uid.CreateCaseRequestBody{
			Type: LpaTypePropertyFinance,
			Donor: uid.DonorDetails{
				Name:     "Jane Smith",
				Dob:      uid.ISODate{Time: date.New("2000", "1", "2").Time()},
				Postcode: "B147ED",
			},
		})

		if err != nil {
			logger.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
