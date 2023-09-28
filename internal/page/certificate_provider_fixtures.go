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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderFixturesData struct {
	App    AppData
	Errors validation.List
}

func CertificateProviderFixtures(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender *ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
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
			"paid",
			"signedByDonor",
			"detailsConfirmed",
		}
		attorneyNames = []Name{
			{Firstnames: "Jessie", Lastname: "Jones"},
			{Firstnames: "Robin", Lastname: "Redcar"},
			{Firstnames: "Leslie", Lastname: "Lewis"},
			{Firstnames: "Ashley", Lastname: "Alwinton"},
			{Firstnames: "Frankie", Lastname: "Fernandes"},
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

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			lpaType  = r.FormValue("lpa-type")
			progress = slices.Index(progressValues, r.FormValue("progress"))
			email    = r.FormValue("email")
			redirect = r.FormValue("redirect")
		)

		if r.Method != http.MethodPost && redirect == "" {
			return tmpl(w, &certificateProviderFixturesData{App: appData})
		}

		var (
			donorSub                     = random.String(16)
			certificateProviderSub       = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: testEmail}); err != nil {
			return err
		}

		lpa, err := donorStore.Create(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
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
		if lpaType == "hw" {
			lpa.Type = LpaTypeHealthWelfare
		}

		lpa.Attorneys = actor.Attorneys{
			Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
		}

		lpa.CertificateProvider = actor.CertificateProvider{
			FirstNames:         "Charlie",
			LastName:           "Cooper",
			Email:              testEmail,
			Mobile:             testMobile,
			Relationship:       actor.Personally,
			RelationshipLength: "gte-2-years",
			CarryOutBy:         actor.Online,
			Address: place.Address{
				Line1:      "5 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}

		if email != "" {
			lpa.CertificateProvider.Email = email
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "paid") {
			lpa.PaymentDetails = append(lpa.PaymentDetails, Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})
			lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
		}
		if progress >= slices.Index(progressValues, "signedByDonor") {
			lpa.SignedAt = time.Now()
		}
		if progress >= slices.Index(progressValues, "detailsConfirmed") {
			certificateProvider.DateOfBirth = date.New("1990", "1", "2")
			certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted
		}

		if err := donorStore.Put(donorCtx, lpa); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if r.FormValue("use-test-code") == "1" {
			useTestCode = true
		}

		if email != "" {
			shareCodeSender.SendCertificateProvider(donorCtx, notify.CertificateProviderInviteEmail, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: appData.Localizer,
			}, true, lpa)

			return AppData{}.Redirect(w, r, nil, Paths.CertificateProviderStart.Format())
		}

		if redirect == "" {
			redirect = Paths.Dashboard.Format()
		} else {
			redirect = "/certificate-provider/" + lpa.ID + redirect
		}

		return AppData{}.Redirect(w, r, nil, redirect)
	}
}
