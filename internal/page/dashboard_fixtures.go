package page

import (
	"encoding/base64"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardFixturesData struct {
	App    AppData
	Errors validation.List
}

func DashboardFixtures(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender *ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
) Handler {
	const (
		testEmail  = "simulate-delivered@notifications.service.gov.uk"
		testMobile = "07700900000"
	)

	type Name struct {
		Firstnames, Lastname string
	}

	attorneyNames := []Name{
		{Firstnames: "Jessie", Lastname: "Jones"},
		{Firstnames: "Robin", Lastname: "Redcar"},
		{Firstnames: "Leslie", Lastname: "Lewis"},
		{Firstnames: "Ashley", Lastname: "Alwinton"},
		{Firstnames: "Frankie", Lastname: "Fernandes"},
	}

	makeAttorney := func(name Name) actor.Attorney {
		return actor.Attorney{
			ID:          name.Firstnames + name.Lastname,
			FirstNames:  name.Firstnames,
			LastName:    name.Lastname,
			Email:       testEmail,
			DateOfBirth: date.New("2000", "1", "2"),
			Address: place.Address{
				Line1:      "2 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}
	}

	makeDonor := func(firstnames, lastname string) actor.Donor {
		return actor.Donor{
			FirstNames: "Sam",
			LastName:   "Smith",
			Address: place.Address{
				Line1:      "1 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
			Email:         testEmail,
			DateOfBirth:   date.New("2000", "1", "2"),
			ThinksCanSign: actor.Yes,
			CanSign:       form.Yes,
		}
	}

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			asDonor               = r.FormValue("asDonor") == "1"
			asAttorney            = r.FormValue("asAttorney") == "1"
			asCertificateProvider = r.FormValue("asCertificateProvider") == "1"
			redirect              = r.FormValue("redirect")
		)

		if r.Method != http.MethodPost && redirect == "" {
			return tmpl(w, &attorneyFixturesData{App: appData})
		}

		var (
			meSub          = random.String(16)
			donorSub       = random.String(16)
			meSessionID    = base64.StdEncoding.EncodeToString([]byte(meSub))
			donorSessionID = base64.StdEncoding.EncodeToString([]byte(donorSub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: meSub, Email: testEmail}); err != nil {
			return err
		}

		if asDonor {
			lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: meSessionID}))
			if err != nil {
				return err
			}

			donorCtx := ContextWithSessionData(r.Context(), &SessionData{SessionID: meSessionID, LpaID: lpa.ID})

			lpa.Donor = makeDonor("Sam", "Smith")
			lpa.Type = LpaTypePropertyFinance

			lpa.Attorneys = actor.Attorneys{
				Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
			}

			if err := donorStore.Put(donorCtx, lpa); err != nil {
				return err
			}
		}

		if asCertificateProvider {
			lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			lpa.Donor = makeDonor("Sam", "Smith")
			lpa.UID = "M-1111-1111-1111"

			if err := donorStore.Put(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID}), lpa); err != nil {
				return err
			}

			certificateProviderCtx := ContextWithSessionData(r.Context(), &SessionData{SessionID: meSessionID, LpaID: lpa.ID})

			certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
			if err != nil {
				return err
			}

			if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
				return err
			}
		}

		if asAttorney {
			lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
			lpa.Donor = makeDonor("Sam", "Smith")
			lpa.Attorneys = actor.Attorneys{
				Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
			}
			lpa.UID = "M-2222-2222-2222"

			if err := donorStore.Put(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID}), lpa); err != nil {
				return err
			}

			attorneyCtx := ContextWithSessionData(r.Context(), &SessionData{SessionID: meSessionID, LpaID: lpa.ID})

			attorney, err := attorneyStore.Create(attorneyCtx, donorSessionID, lpa.Attorneys.Attorneys[0].ID, false, false)
			if err != nil {
				return err
			}

			if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
				return err
			}
		}

		return AppData{}.Redirect(w, r, nil, Paths.Dashboard.Format())
	}
}
