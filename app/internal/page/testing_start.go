package page

import (
	"context"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
)

func TestingStart(store sesh.Store, donorStore DonorStore, randomString func(int) string, shareCodeSender shareCodeSender, localizer Localizer, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore, logger *logging.Logger, now func() time.Time) http.HandlerFunc {
	const (
		testEmail  = "simulate-delivered@notifications.service.gov.uk"
		testMobile = "07700900000"
	)

	var (
		attorneyNames            = []string{"John", "Joan", "Johan", "Jilly", "James"}
		replacementAttorneyNames = []string{"Jane", "Jorge", "Jackson", "Jacob", "Joshua"}
		peopleToNotifyNames      = []string{"Joanna", "Jonathan", "Julian", "Jayden", "Juniper"}
	)

	type lpaOptions struct {
		hasDonorDetails            bool
		lpaType                    string
		attorneys                  int
		howAttorneysAct            string
		replacementAttorneys       int
		howReplacementAttorneysAct string
		hasWhenCanBeUsed           bool
		hasRestrictions            bool
		hasCertificateProvider     bool
		peopleToNotify             int
		checked                    bool
		paid                       bool
		idConfirmedAndSigned       bool
		submitted                  bool
		attorneyEmail              string
		replacementAttorneyEmail   string
		certificateProviderEmail   string
	}

	makeDonor := func() actor.Donor {
		return actor.Donor{
			FirstNames: "Jamie",
			LastName:   "Smith",
			Address: place.Address{
				Line1:      "1 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
			Email:       testEmail,
			DateOfBirth: date.New("2000", "1", "2"),
		}
	}

	makeAttorney := func(firstNames string) actor.Attorney {
		return actor.Attorney{
			ID:          firstNames + "Smith",
			FirstNames:  firstNames,
			LastName:    "Smith",
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

	makePersonToNotify := func(firstNames string) actor.PersonToNotify {
		return actor.PersonToNotify{
			ID:         firstNames + "Smith",
			FirstNames: firstNames,
			LastName:   "Smith",
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

	makeCertificateProvider := func(firstNames string) actor.CertificateProvider {
		return actor.CertificateProvider{
			FirstNames:         firstNames,
			LastName:           "Jones",
			Email:              testEmail,
			Mobile:             testMobile,
			Relationship:       actor.Personally,
			RelationshipLength: "gte-2-years",
			CarryOutBy:         "paper",
			Address: place.Address{
				Line1:      "5 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
		}
	}

	addAttorneys := func(lpa *Lpa, count int) {
		if count > len(attorneyNames) {
			count = len(attorneyNames)
		}

		for _, name := range attorneyNames[:count] {
			lpa.Attorneys = append(lpa.Attorneys, makeAttorney(name))
		}

		if count > 1 {
			lpa.AttorneyDecisions.How = actor.JointlyAndSeverally
		}

		lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
	}

	addReplacementAttorneys := func(lpa *Lpa, count int) {
		if count > len(replacementAttorneyNames) {
			count = len(replacementAttorneyNames)
		}

		for _, name := range replacementAttorneyNames[:count] {
			lpa.ReplacementAttorneys = append(lpa.ReplacementAttorneys, makeAttorney(name))
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
				lpa.Tasks.YourDetails = actor.TaskCompleted
			}

			if opts.lpaType == "hw" {
				lpa.Type = LpaTypeHealthWelfare
			}

			if opts.attorneys > 0 {
				addAttorneys(lpa, opts.attorneys)
			}

			if opts.attorneys == -1 {
				addAttorneys(lpa, 2)
				lpa.Attorneys[0].ID = "with-address"
				lpa.Attorneys[1].ID = "without-address"
				lpa.Attorneys[1].Address = place.Address{}

				lpa.ReplacementAttorneys = lpa.Attorneys
				lpa.Type = LpaTypePropertyFinance
				lpa.WhenCanTheLpaBeUsed = CanBeUsedWhenRegistered

				lpa.AttorneyDecisions.How = actor.JointlyAndSeverally

				lpa.WantReplacementAttorneys = form.Yes
				lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
				lpa.HowShouldReplacementAttorneysStepIn = ReplacementAttorneysStepInWhenOneCanNoLongerAct

				lpa.Tasks.ChooseAttorneys = actor.TaskInProgress
				lpa.Tasks.ChooseReplacementAttorneys = actor.TaskInProgress
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
				addReplacementAttorneys(lpa, opts.replacementAttorneys)
			}

			if opts.replacementAttorneys == -1 {
				addReplacementAttorneys(lpa, 2)
				lpa.ReplacementAttorneys[0].ID = "with-address"
				lpa.ReplacementAttorneys[1].ID = "without-address"
				lpa.ReplacementAttorneys[1].Address = place.Address{}

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
				lpa.WhenCanTheLpaBeUsed = CanBeUsedWhenRegistered
				lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
				lpa.LifeSustainingTreatmentOption = LifeSustainingTreatmentOptionA
				lpa.Tasks.LifeSustainingTreatment = actor.TaskCompleted
			}

			if opts.hasRestrictions {
				lpa.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
				lpa.Tasks.Restrictions = actor.TaskCompleted
			}

			if opts.hasCertificateProvider {
				lpa.CertificateProvider = makeCertificateProvider("Jessie")
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
				lpa.Checked = true
				lpa.HappyToShare = true
				lpa.Tasks.CheckYourLpa = actor.TaskCompleted
			}

			if opts.paid {
				ref := randomString(12)
				sesh.SetPayment(store, r, w, &sesh.PaymentSession{PaymentID: ref})

				lpa.PaymentDetails = PaymentDetails{
					PaymentReference: ref,
					PaymentId:        ref,
				}
				lpa.Tasks.PayForLpa = actor.TaskCompleted
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

			if opts.attorneyEmail != "" {
				lpa.Attorneys[0].Email = opts.attorneyEmail
			}

			if opts.replacementAttorneyEmail != "" {
				lpa.ReplacementAttorneys[0].Email = opts.replacementAttorneyEmail
			}

			if err := donorStore.Put(ctx, lpa); err != nil {
				logger.Print("putting lpa ", err)
			}

			return lpa
		}

		var (
			completeLpa                = r.FormValue("lpa.complete") != ""
			cookiesAccepted            = r.FormValue("cookiesAccepted") != ""
			useTestShareCode           = r.FormValue("useTestShareCode") != ""
			withShareCodeSession       = r.FormValue("withShareCodeSession") != ""
			startCpFlowDonorHasPaid    = r.FormValue("startCpFlowDonorHasPaid") != ""
			startCpFlowDonorHasNotPaid = r.FormValue("startCpFlowDonorHasNotPaid") != ""
			asCertificateProvider      = r.FormValue("certificateProviderProvided")
			fresh                      = r.FormValue("fresh") != ""
			asAttorney                 = r.FormValue("attorneyProvided") != ""
			asReplacementAttorney      = r.FormValue("replacementAttorneyProvided") != ""
			sendAttorneyShare          = r.FormValue("sendAttorneyShare") != ""
			redirect                   = r.FormValue("redirect")
			paymentComplete            = r.FormValue("lpa.paid") != ""
		)

		completeSectionOne := completeLpa || startCpFlowDonorHasNotPaid || startCpFlowDonorHasPaid || paymentComplete

		lpa := buildLpa(ContextWithSessionData(r.Context(), &SessionData{SessionID: donorSessionID}), lpaOptions{
			hasDonorDetails:            r.FormValue("lpa.yourDetails") != "" || completeSectionOne,
			lpaType:                    r.FormValue("lpa.type"),
			attorneys:                  parseCount(r.FormValue("lpa.attorneys"), completeSectionOne),
			howAttorneysAct:            r.FormValue("lpa.attorneysAct"),
			replacementAttorneys:       parseCount(r.FormValue("lpa.replacementAttorneys"), completeSectionOne),
			howReplacementAttorneysAct: r.FormValue("lpa.replacementAttorneysAct"),
			hasWhenCanBeUsed:           r.FormValue("lpa.chooseWhenCanBeUsed") != "" || completeSectionOne,
			hasRestrictions:            r.FormValue("lpa.restrictions") != "" || completeSectionOne,
			hasCertificateProvider:     r.FormValue("lpa.certificateProvider") != "" || completeSectionOne,
			peopleToNotify:             parseCount(r.FormValue("lpa.peopleToNotify"), completeSectionOne),
			checked:                    r.FormValue("lpa.checkAndSend") != "" || completeSectionOne,
			paid:                       paymentComplete || startCpFlowDonorHasPaid || completeLpa,
			idConfirmedAndSigned:       r.FormValue("lpa.confirmIdentityAndSign") != "" || completeLpa,
			submitted:                  r.FormValue("lpa.signedByDonor") != "",
			certificateProviderEmail:   r.FormValue("lpa.certificateProviderEmail"),
			attorneyEmail:              r.FormValue("lpa.attorneyEmail"),
			replacementAttorneyEmail:   r.FormValue("lpa.replacementAttorneyEmail"),
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

			_, err := attorneyStore.Create(currentCtx, attorneySessionID, lpa.Attorneys[0].ID, false)
			if err != nil {
				logger.Print("asAttorney:", err)
			}
		}

		if asReplacementAttorney {
			_, err := attorneyStore.Create(attorneyCtx, attorneySessionID, lpa.ReplacementAttorneys[0].ID, true)
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

		AppData{}.Redirect(w, r.WithContext(donorCtx), lpa, redirect)
	}
}
