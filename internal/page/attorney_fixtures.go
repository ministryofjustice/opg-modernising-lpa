package page

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type attorneyFixturesData struct {
	App    AppData
	Errors validation.List
}

func AttorneyFixtures(
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

	var (
		progressValues = []string{
			"signedByCertificateProvider",
			"signedByAttorney",
			"submitted",
			"registered",
		}
		attorneyNames = []Name{
			{Firstnames: "Jessie", Lastname: "Jones"},
			{Firstnames: "Robin", Lastname: "Redcar"},
			{Firstnames: "Leslie", Lastname: "Lewis"},
			{Firstnames: "Ashley", Lastname: "Alwinton"},
			{Firstnames: "Frankie", Lastname: "Fernandes"},
		}
		replacementAttorneyNames = []Name{
			{Firstnames: "Blake", Lastname: "Buckley"},
			{Firstnames: "Taylor", Lastname: "Thompson"},
			{Firstnames: "Marley", Lastname: "Morris"},
			{Firstnames: "Alex", Lastname: "Abbott"},
			{Firstnames: "Billie", Lastname: "Blair"},
		}
	)

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

	makeTrustCorporation := func(name string) actor.TrustCorporation {
		return actor.TrustCorporation{
			Name:          name,
			CompanyNumber: "555555555",
			Email:         testEmail,
			Address: place.Address{
				Line1:      "2 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}
	}

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			isReplacement      = r.FormValue("is-replacement") == "1"
			isTrustCorporation = r.FormValue("is-trust-corporation") == "1"
			lpaType            = r.FormValue("lpa-type")
			progress           = slices.Index(progressValues, r.FormValue("progress"))
			email              = r.FormValue("email")
			redirect           = r.FormValue("redirect")
		)

		if r.Method != http.MethodPost && redirect == "" {
			return tmpl(w, &attorneyFixturesData{App: appData})
		}

		if lpaType == "hw" && isTrustCorporation {
			return tmpl(w, &attorneyFixturesData{App: appData, Errors: validation.With("", validation.CustomError{Label: "Can't add a trust corporation to a personal welfare LPA"})})
		}

		var (
			donorSub                     = random.String(16)
			attorneySub                  = random.String(16)
			certificateProviderSub       = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(attorneySub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: attorneySub, Email: testEmail}); err != nil {
			return err
		}

		lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
			attorneyCtx            = ContextWithSessionData(r.Context(), &SessionData{SessionID: attorneySessionID, LpaID: lpa.ID})
		)

		lpa.Donor = actor.Donor{
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
		lpa.Type = LpaTypePropertyFinance
		if lpaType == "hw" && !isTrustCorporation {
			lpa.Type = LpaTypeHealthWelfare
		}

		lpa.Attorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(attorneyNames[0])},
			TrustCorporation: makeTrustCorporation("First Choice Trust Corporation Ltd."),
		}
		lpa.ReplacementAttorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(replacementAttorneyNames[0])},
			TrustCorporation: makeTrustCorporation("Second Choice Trust Corporation Ltd."),
		}

		if email != "" {
			if isTrustCorporation {
				if isReplacement {
					lpa.ReplacementAttorneys.TrustCorporation.Email = email
				} else {
					lpa.Attorneys.TrustCorporation.Email = email
				}
			}
			if isReplacement {
				lpa.ReplacementAttorneys.Attorneys[0].Email = email
			} else {
				lpa.Attorneys.Attorneys[0].Email = email
			}
		}

		var attorneyID string
		if !isTrustCorporation {
			if isReplacement {
				attorneyID = lpa.ReplacementAttorneys.Attorneys[0].ID
			} else {
				attorneyID = lpa.Attorneys.Attorneys[0].ID
			}
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		certificateProvider.Certificate = actor.Certificate{Agreed: time.Now()}

		attorney, err := attorneyStore.Create(attorneyCtx, donorSessionID, attorneyID, isReplacement, isTrustCorporation)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			lpa.SignedAt = time.Now()
		}
		if progress >= slices.Index(progressValues, "signedByAttorney") {
			attorney.Mobile = testMobile
			attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
			attorney.Tasks.ReadTheLpa = actor.TaskCompleted
			attorney.Tasks.SignTheLpa = actor.TaskCompleted

			if isTrustCorporation {
				attorney.WouldLikeSecondSignatory = form.No
				attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
					FirstNames:        "A",
					LastName:          "Sign",
					ProfessionalTitle: "Assistant to the signer",
					Confirmed:         time.Now(),
				}}
			} else {
				attorney.Confirmed = time.Now()
			}
		}
		if progress >= slices.Index(progressValues, "submitted") {
			lpa.SubmittedAt = time.Now()
		}
		if progress >= slices.Index(progressValues, "registered") {
			lpa.RegisteredAt = time.Now()
		}

		if err := donorStore.Put(donorCtx, lpa); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}
		if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if r.FormValue("use-test-code") == "1" {
			useTestCode = true
		}

		if email != "" {
			shareCodeSender.SendAttorneys(donorCtx, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: appData.Localizer,
			}, lpa)

			return AppData{}.Redirect(w, r, nil, Paths.Attorney.Start.Format())
		}

		if redirect == "" {
			redirect = Paths.Dashboard.Format()
		} else {
			redirect = "/attorney/" + lpa.ID + redirect
		}

		return AppData{}.Redirect(w, r, nil, redirect)
	}
}
