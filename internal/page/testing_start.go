package page

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func TestingStart(store sesh.Store, donorStore DonorStore, randomString func(int) string, shareCodeSender shareCodeSender, localizer Localizer, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore, logger *logging.Logger, now func() time.Time) http.HandlerFunc {
	const (
		testEmail  = "simulate-delivered@notifications.service.gov.uk"
		testMobile = "07700900000"
	)

	type Name struct {
		Firstnames, Lastname string
	}

	var (
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
		peopleToNotifyNames = []Name{
			{Firstnames: "Jordan", Lastname: "Jefferson"},
			{Firstnames: "Danni", Lastname: "Davies"},
			{Firstnames: "Bobbie", Lastname: "Bones"},
			{Firstnames: "Ally", Lastname: "Avery"},
			{Firstnames: "Deva", Lastname: "Dankar"},
		}
	)

	type lpaOptions struct {
		hasDonorDetails                  bool
		lpaType                          string
		attorneys                        int
		trustCorporation                 string
		replacementTrustCorporation      string
		howAttorneysAct                  string
		replacementAttorneys             int
		howReplacementAttorneysAct       string
		hasWhenCanBeUsed                 bool
		hasRestrictions                  bool
		hasCertificateProvider           bool
		peopleToNotify                   int
		checked                          bool
		paid                             bool
		idConfirmedAndSigned             bool
		submitted                        bool
		attorneyEmail                    string
		replacementAttorneyEmail         string
		certificateProviderEmail         string
		trustCorporationEmail            string
		replacementTrustCorporationEmail string
		certificateProviderActOnline     bool
	}

	makeDonor := func() actor.Donor {
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

	makePersonToNotify := func(name Name) actor.PersonToNotify {
		return actor.PersonToNotify{
			ID:         name.Firstnames + name.Lastname,
			FirstNames: name.Firstnames,
			LastName:   name.Lastname,
			Email:      testEmail,
			Address: place.Address{
				Line1:      "4 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}
	}

	makeCertificateProvider := func(carryOutBy actor.CertificateProviderCarryOutBy) actor.CertificateProvider {
		return actor.CertificateProvider{
			FirstNames:         "Charlie",
			LastName:           "Cooper",
			Email:              testEmail,
			Mobile:             testMobile,
			Relationship:       actor.Personally,
			RelationshipLength: "gte-2-years",
			CarryOutBy:         carryOutBy,
			Address: place.Address{
				Line1:      "5 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}
	}

	addAttorneys := func(opts lpaOptions, lpa *Lpa) {
		count := opts.attorneys

		if count > len(attorneyNames) {
			count = len(attorneyNames)
		}

		for i, name := range attorneyNames[:count] {
			a := makeAttorney(name)
			if i == 0 && opts.attorneyEmail != "" {
				a.Email = opts.attorneyEmail
			}

			lpa.Attorneys.Attorneys = append(lpa.Attorneys.Attorneys, a)
		}

		if count > 1 {
			lpa.AttorneyDecisions.How = actor.JointlyAndSeverally
		}

		lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
	}

	addReplacementAttorneys := func(opts lpaOptions, lpa *Lpa) {
		count := opts.replacementAttorneys

		if count > len(replacementAttorneyNames) {
			count = len(replacementAttorneyNames)
		}

		for i, name := range replacementAttorneyNames[:count] {
			a := makeAttorney(name)
			if i == 0 && opts.replacementAttorneyEmail != "" {
				a.Email = opts.replacementAttorneyEmail
			}

			lpa.ReplacementAttorneys.Attorneys = append(lpa.ReplacementAttorneys.Attorneys, a)
		}

		if count > 1 {
			lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
			lpa.HowShouldReplacementAttorneysStepIn = ReplacementAttorneysStepInWhenOneCanNoLongerAct
		}

		lpa.WantReplacementAttorneys = form.Yes
		lpa.Tasks.ChooseReplacementAttorneys = actor.TaskCompleted
	}

	addPeopleToNotify := func(lpa *Lpa, count int) {
		if count > len(peopleToNotifyNames) {
			count = len(peopleToNotifyNames)
		}

		for _, name := range peopleToNotifyNames[:count] {
			lpa.PeopleToNotify = append(lpa.PeopleToNotify, makePersonToNotify(name))
		}

		lpa.DoYouWantToNotifyPeople = form.Yes
		lpa.Tasks.PeopleToNotify = actor.TaskCompleted
	}

	parseCount := func(s string, complete bool) int {
		switch s {
		case "":
			if complete {
				return 2
			}

			return 0
		case "incomplete":
			return -1
		default:
			if count, err := strconv.Atoi(s); err == nil {
				return count
			}

			return 2
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			donorSub                     = randomString(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSub       = randomString(16)
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySub                  = randomString(16)
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(attorneySub))
		)

		buildLpa := func(ctx context.Context, opts lpaOptions) *Lpa {
			lpa, err := donorStore.Create(ctx)

			if err != nil {
				logger.Print("creating lpa ", err)
			}

			if opts.hasDonorDetails {
				lpa.Donor = makeDonor()
				lpa.WhoFor = "me"
				lpa.Type = LpaTypePropertyFinance
				lpa.UID = random.UuidString()
				lpa.Tasks.YourDetails = actor.TaskCompleted
			}

			if opts.lpaType == "hw" {
				lpa.Type = LpaTypeHealthWelfare
			}

			if opts.attorneys > 0 {
				addAttorneys(opts, lpa)
			}

			if opts.attorneys == -1 {
				attorney1 := makeAttorney(attorneyNames[0])
				attorney1.ID = "with-address"

				attorney2 := makeAttorney(attorneyNames[1])
				attorney2.ID = "without-address"
				attorney2.Address = place.Address{}

				lpa.Attorneys.Attorneys = []actor.Attorney{attorney1, attorney2}
				lpa.AttorneyDecisions.How = actor.JointlyAndSeverally

				lpa.ReplacementAttorneys = lpa.Attorneys
				lpa.Type = LpaTypePropertyFinance
				lpa.WhenCanTheLpaBeUsed = CanBeUsedWhenHasCapacity

				lpa.WantReplacementAttorneys = form.Yes
				lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
				lpa.HowShouldReplacementAttorneysStepIn = ReplacementAttorneysStepInWhenOneCanNoLongerAct

				lpa.Tasks.ChooseAttorneys = actor.TaskInProgress
				lpa.Tasks.ChooseReplacementAttorneys = actor.TaskInProgress
			}

			switch opts.trustCorporation {
			case "incomplete":
				lpa.Attorneys.TrustCorporation = actor.TrustCorporation{Name: "My company"}
			case "complete":
				lpa.Attorneys.TrustCorporation = actor.TrustCorporation{
					Name:          "My company",
					CompanyNumber: "555555555",
					Email:         testEmail,
					Address:       place.Address{Line1: "123 Fake Street", Postcode: "FF1 1FF"},
				}
			}

			if opts.trustCorporationEmail != "" {
				lpa.Attorneys.TrustCorporation.Email = opts.trustCorporationEmail
			}

			switch opts.replacementTrustCorporation {
			case "incomplete":
				lpa.ReplacementAttorneys.TrustCorporation = actor.TrustCorporation{Name: "My company"}
			case "complete":
				lpa.ReplacementAttorneys.TrustCorporation = actor.TrustCorporation{
					Name:          "My company",
					CompanyNumber: "555555555",
					Email:         testEmail,
					Address:       place.Address{Line1: "123 Fake Street", Postcode: "FF1 1FF"},
				}
			}

			if opts.replacementTrustCorporationEmail != "" {
				lpa.ReplacementAttorneys.TrustCorporation.Email = opts.replacementTrustCorporationEmail
			}

			if opts.howAttorneysAct != "" {
				act, err := actor.ParseAttorneysAct(opts.howAttorneysAct)
				if err != nil {
					act = actor.JointlyForSomeSeverallyForOthers
				}

				lpa.AttorneyDecisions.How = act
				if act == actor.JointlyForSomeSeverallyForOthers {
					lpa.AttorneyDecisions.Details = "some details"
				}
			}

			if opts.replacementAttorneys > 0 {
				addReplacementAttorneys(opts, lpa)
			}

			if opts.replacementAttorneys == -1 {
				attorney1 := makeAttorney(attorneyNames[0])
				attorney1.ID = "with-address"

				attorney2 := makeAttorney(attorneyNames[1])
				attorney2.ID = "without-address"
				attorney2.Address = place.Address{}

				lpa.ReplacementAttorneys.Attorneys = []actor.Attorney{attorney1, attorney2}
				lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally

				lpa.ReplacementAttorneys = lpa.Attorneys
				lpa.WantReplacementAttorneys = form.Yes
				lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
				lpa.Tasks.ChooseReplacementAttorneys = actor.TaskInProgress
			}

			if opts.howReplacementAttorneysAct != "" {
				act, err := actor.ParseAttorneysAct(opts.howReplacementAttorneysAct)
				if err != nil {
					act = actor.JointlyForSomeSeverallyForOthers
				}

				lpa.ReplacementAttorneyDecisions.How = act
				if act == actor.JointlyForSomeSeverallyForOthers {
					lpa.ReplacementAttorneyDecisions.Details = "some details"
				}
			}

			if opts.hasWhenCanBeUsed {
				lpa.WhenCanTheLpaBeUsed = CanBeUsedWhenHasCapacity
				lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
				lpa.LifeSustainingTreatmentOption = LifeSustainingTreatmentOptionA
				lpa.Tasks.LifeSustainingTreatment = actor.TaskCompleted
			}

			if opts.hasRestrictions {
				lpa.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
				lpa.Tasks.Restrictions = actor.TaskCompleted
			}

			if opts.hasCertificateProvider {
				carryOutBy := actor.Paper

				if opts.certificateProviderActOnline {
					carryOutBy = actor.Online
				}

				lpa.CertificateProvider = makeCertificateProvider(carryOutBy)
				lpa.Tasks.CertificateProvider = actor.TaskCompleted
			}

			if opts.peopleToNotify > 0 {
				addPeopleToNotify(lpa, opts.peopleToNotify)
			}

			if opts.peopleToNotify == -1 {
				addPeopleToNotify(lpa, 1)

				joanna := lpa.PeopleToNotify[0]
				joanna.Address = place.Address{}
				lpa.PeopleToNotify = actor.PeopleToNotify{
					joanna,
				}

				lpa.Tasks.PeopleToNotify = actor.TaskInProgress
			}

			if opts.checked {
				lpa.CheckedAndHappy = true
				lpa.Tasks.CheckYourLpa = actor.TaskCompleted
			}

			if opts.paid {
				ref := randomString(12)
				sesh.SetPayment(store, r, w, &sesh.PaymentSession{PaymentID: ref})

				lpa.PaymentDetails = append(lpa.PaymentDetails, Payment{
					PaymentReference: ref,
					PaymentId:        ref,
				})
				lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
			}

			if opts.idConfirmedAndSigned {
				lpa.DonorIdentityUserData = identity.UserData{
					OK:          true,
					Provider:    identity.OneLogin,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FirstNames:  "Jamie",
					LastName:    "Smith",
				}

				lpa.WantToApplyForLpa = true
				lpa.WantToSignLpa = true
				lpa.Submitted = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
				lpa.CPWitnessCodeValidated = true
				lpa.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
			}

			if opts.submitted {
				lpa.Submitted = now()
			}

			if opts.certificateProviderEmail != "" {
				lpa.CertificateProvider.Email = opts.certificateProviderEmail
			}

			if err := donorStore.Put(ctx, lpa); err != nil {
				logger.Print("putting lpa ", err)
			}

			return lpa
		}

		var (
			completeLpa                  = r.FormValue("lpa.complete") != ""
			cookiesAccepted              = r.FormValue("cookiesAccepted") != ""
			useTestShareCode             = r.FormValue("useTestShareCode") != ""
			withShareCodeSession         = r.FormValue("withShareCodeSession") != ""
			startCpFlowDonorHasPaid      = r.FormValue("startCpFlowDonorHasPaid") != ""
			startCpFlowDonorHasNotPaid   = r.FormValue("startCpFlowDonorHasNotPaid") != ""
			asCertificateProvider        = r.FormValue("asCertificateProvider")
			cpConfirmYourDetailsComplete = r.FormValue("cp.confirmYourDetails") != ""
			fresh                        = r.FormValue("fresh") != ""
			asAttorney                   = r.FormValue("attorneyProvided") != ""
			asReplacementAttorney        = r.FormValue("replacementAttorneyProvided") != ""
			sendAttorneyShare            = r.FormValue("sendAttorneyShare") != ""
			redirect                     = r.FormValue("redirect")
			paymentComplete              = r.FormValue("lpa.paid") != ""
		)

		completeSectionOne := completeLpa || startCpFlowDonorHasNotPaid || startCpFlowDonorHasPaid || paymentComplete

		lpa := buildLpa(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}), lpaOptions{
			hasDonorDetails:                  r.FormValue("lpa.yourDetails") != "" || completeSectionOne,
			lpaType:                          r.FormValue("lpa.type"),
			attorneys:                        parseCount(r.FormValue("lpa.attorneys"), completeSectionOne),
			trustCorporation:                 r.FormValue("lpa.trustCorporation"),
			replacementTrustCorporation:      r.FormValue("lpa.replacementTrustCorporation"),
			howAttorneysAct:                  r.FormValue("lpa.attorneysAct"),
			replacementAttorneys:             parseCount(r.FormValue("lpa.replacementAttorneys"), completeSectionOne),
			howReplacementAttorneysAct:       r.FormValue("lpa.replacementAttorneysAct"),
			hasWhenCanBeUsed:                 r.FormValue("lpa.chooseWhenCanBeUsed") != "" || completeSectionOne,
			hasRestrictions:                  r.FormValue("lpa.restrictions") != "" || completeSectionOne,
			hasCertificateProvider:           r.FormValue("lpa.certificateProvider") != "" || completeSectionOne,
			certificateProviderActOnline:     r.FormValue("lpa.certificateProviderActOnline") != "",
			peopleToNotify:                   parseCount(r.FormValue("lpa.peopleToNotify"), completeSectionOne),
			checked:                          r.FormValue("lpa.checkAndSend") != "" || completeSectionOne,
			paid:                             paymentComplete || startCpFlowDonorHasPaid || completeLpa,
			idConfirmedAndSigned:             r.FormValue("lpa.confirmIdentityAndSign") != "" || completeLpa,
			submitted:                        r.FormValue("lpa.signedByDonor") != "",
			certificateProviderEmail:         r.FormValue("lpa.certificateProviderEmail"),
			attorneyEmail:                    r.FormValue("lpa.attorneyEmail"),
			replacementAttorneyEmail:         r.FormValue("lpa.replacementAttorneyEmail"),
			trustCorporationEmail:            r.FormValue("lpa.trustCorporationEmail"),
			replacementTrustCorporationEmail: r.FormValue("lpa.replacementTrustCorporationEmail"),
		})

		// These contexts act on the same LPA for different actors
		var (
			donorCtx               = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
			attorneyCtx            = ContextWithSessionData(r.Context(), &SessionData{SessionID: attorneySessionID, LpaID: lpa.ID})
		)

		// loginAs controls which actor we will be pretending to be for the LPA
		switch r.FormValue("loginAs") {
		case "attorney":
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: attorneySub, Email: testEmail})
			if redirect != "" {
				redirect = "/attorney/" + lpa.ID + redirect
			}
		case "certificate-provider":
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: testEmail})
			if redirect != "" {
				redirect = "/certificate-provider/" + lpa.ID + redirect
			}
		default:
			_ = sesh.SetLoginSession(store, r, w, &sesh.LoginSession{Sub: donorSub, Email: testEmail})
			if redirect != "" &&
				redirect != Paths.Start.Format() &&
				redirect != Paths.Dashboard.Format() &&
				redirect != Paths.CertificateProviderStart.Format() &&
				redirect != Paths.Attorney.Start.Format() {
				redirect = "/lpa/" + lpa.ID + redirect
			}
		}

		if cookiesAccepted {
			http.SetCookie(w, &http.Cookie{
				Name:   "cookies-consent",
				Value:  "accept",
				MaxAge: 365 * 24 * 60 * 60,
				Path:   "/",
			})
		}

		if useTestShareCode {
			shareCodeSender.UseTestCode()
		}

		if withShareCodeSession {
			sesh.SetShareCode(store, r, w, &sesh.ShareCodeSession{LpaID: lpa.ID, Identity: false})
		}

		if startCpFlowDonorHasPaid || startCpFlowDonorHasNotPaid {
			shareCodeSender.SendCertificateProvider(donorCtx, notify.CertificateProviderInviteEmail, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, false, lpa)

			redirect = Paths.CertificateProviderStart.Format()
		}

		if asCertificateProvider != "" {
			currentCtx := certificateProviderCtx

			// "fresh=1" causes an LPA to be created for the certificate provider (as
			// the donor), we then link this to the donor so they are both each
			// other's certificate provider.
			if fresh {
				lpa := buildLpa(certificateProviderCtx, lpaOptions{
					hasDonorDetails:        true,
					attorneys:              2,
					replacementAttorneys:   2,
					hasWhenCanBeUsed:       true,
					hasRestrictions:        true,
					hasCertificateProvider: true,
					peopleToNotify:         2,
					checked:                true,
					paid:                   true,
					idConfirmedAndSigned:   true,
				})

				currentCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			}

			certificateProvider, err := certificateProviderStore.Create(currentCtx, certificateProviderSessionID)
			if err != nil {
				logger.Print("asCertificateProvider creating CP ", err)
			}

			certificateProvider.IdentityUserData = identity.UserData{
				OK:         true,
				Provider:   identity.OneLogin,
				FirstNames: "Jessie",
				LastName:   "Jones",
			}

			if asCertificateProvider == "certified" {
				certificateProvider.Mobile = testMobile
				certificateProvider.Email = testEmail
				certificateProvider.Certificate = actor.Certificate{
					AgreeToStatement: true,
					Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				}
			}

			if cpConfirmYourDetailsComplete {
				certificateProvider.Mobile = testMobile
				certificateProvider.Email = testEmail
				certificateProvider.DateOfBirth = date.New("2000", "1", "2")
				certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted
			}

			if err := certificateProviderStore.Put(currentCtx, certificateProvider); err != nil {
				logger.Print("provideCertificate putting CP ", err)
			}
		}

		if asAttorney {
			currentCtx := attorneyCtx

			if fresh {
				lpa := buildLpa(attorneyCtx, lpaOptions{
					hasDonorDetails:        true,
					attorneys:              2,
					replacementAttorneys:   2,
					hasWhenCanBeUsed:       true,
					hasRestrictions:        true,
					hasCertificateProvider: true,
					peopleToNotify:         2,
					checked:                true,
					paid:                   true,
					idConfirmedAndSigned:   true,
				})

				currentCtx = ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			}

			id := lpa.Attorneys.Attorneys[0].ID

			_, err := attorneyStore.Create(currentCtx, attorneySessionID, id, false)
			if err != nil {
				logger.Print("asAttorney:", err)
			}
		}

		if asReplacementAttorney {
			id := lpa.ReplacementAttorneys.Attorneys[0].ID

			_, err := attorneyStore.Create(attorneyCtx, attorneySessionID, id, true)
			if err != nil {
				logger.Print("asReplacementAttorney:", err)
			}
		}

		if sendAttorneyShare {
			shareCodeSender.SendAttorneys(donorCtx, AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: localizer,
			}, lpa)
		}

		random.UseTestCode = true

		lang := localize.En
		if r.FormValue("lang") == "cy" {
			lang = localize.Cy
		}

		AppData{Lang: lang}.Redirect(w, r.WithContext(donorCtx), lpa, redirect)
	}
}
