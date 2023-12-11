package fixtures

import (
	"context"
	"encoding/base64"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DonorStore interface {
	Create(context.Context) (*actor.DonorProvidedDetails, error)
	Put(context.Context, *actor.DonorProvidedDetails) error
}

type CertificateProviderStore interface {
	Create(context.Context, string) (*actor.CertificateProviderProvidedDetails, error)
	Put(context.Context, *actor.CertificateProviderProvidedDetails) error
}

type AttorneyStore interface {
	Create(context.Context, string, string, bool, bool) (*actor.AttorneyProvidedDetails, error)
	Put(context.Context, *actor.AttorneyProvidedDetails) error
}

func Attorney(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	oneloginClient *onelogin.Client,
	dynamodbClient DynamoClient,
) page.Handler {
	progressValues := []string{
		"signedByCertificateProvider",
		"signedByAttorney",
		"signedByAllAttorneys",
		"submitted",
		"withdrawn",
		"registered",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			isReplacement      = r.FormValue("is-replacement") == "1"
			isTrustCorporation = r.FormValue("is-trust-corporation") == "1"
			lpaType            = r.FormValue("lpa-type")
			progress           = slices.Index(progressValues, r.FormValue("progress"))
			email              = r.FormValue("email")
			redirect           = r.FormValue("redirect")
			withLpaUID         = r.FormValue("loginWithLpaUID")
		)

		if withLpaUID != "" {
			notFoundError := validation.With("loginWithLpaUID", validation.CustomError{Label: "Attorney not found for LPA UID " + withLpaUID})

			var donor actor.DonorProvidedDetails
			if err := dynamodbClient.OneByUID(context.Background(), withLpaUID, &donor); err != nil {
				return tmpl(w, &fixturesData{App: appData, Errors: notFoundError})
			}

			var links []*lpaLink
			if err := dynamodbClient.AllByPartialSk(context.Background(), donor.PK, "#SUB#", &links); err != nil {
				return tmpl(w, &fixturesData{App: appData, Errors: notFoundError})
			}

			sub := ""
			for _, link := range links {
				if link.ActorType == actor.TypeAttorney {
					decodedSub, err := base64.StdEncoding.DecodeString(strings.Split(link.SK, "#SUB#")[1])
					if err != nil {
						return tmpl(w, &fixturesData{App: appData, Errors: notFoundError})
					}

					sub = string(decodedSub)
					break
				}
			}

			if sub == "" {
				return tmpl(w, &fixturesData{App: appData, Errors: notFoundError})
			}

			state := "abc123"
			nonce := "xyz456"

			authCodeURL, err := oneloginClient.AuthCodeURL(state, nonce, localize.En.String(), false)
			if err != nil {
				return tmpl(w, &fixturesData{App: appData, Errors: notFoundError})
			}

			if err := sesh.SetOneLogin(sessionStore, r, w, &sesh.OneLoginSession{
				State:    state,
				Nonce:    nonce,
				Redirect: page.Paths.Attorney.LoginCallback.Format(),
			}); err != nil {
				return nil
			}

			http.Redirect(w, r, authCodeURL+"&sub="+sub, http.StatusFound)

			return nil
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData})
		}

		if lpaType == "hw" && isTrustCorporation {
			return tmpl(w, &fixturesData{App: appData, Errors: validation.With("", validation.CustomError{Label: "Can't add a trust corporation to a personal welfare LPA"})})
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

		donor, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donor.LpaID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: donor.LpaID})
			attorneyCtx            = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: attorneySessionID, LpaID: donor.LpaID})
		)

		donor.Donor = actor.Donor{
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

		donor.LpaUID = makeUID()
		if lpaType == "hw" && !isTrustCorporation {
			donor.Type = actor.LpaTypeHealthWelfare
			donor.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
			donor.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
		} else {
			donor.Type = actor.LpaTypePropertyFinance
			donor.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
		}

		donor.CertificateProvider = makeCertificateProvider()

		donor.Attorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(attorneyNames[0])},
			TrustCorporation: makeTrustCorporation("First Choice Trust Corporation Ltd."),
		}
		donor.ReplacementAttorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(replacementAttorneyNames[0])},
			TrustCorporation: makeTrustCorporation("Second Choice Trust Corporation Ltd."),
		}

		if email != "" {
			if isTrustCorporation {
				if isReplacement {
					donor.ReplacementAttorneys.TrustCorporation.Email = email
				} else {
					donor.Attorneys.TrustCorporation.Email = email
				}
			}
			if isReplacement {
				donor.ReplacementAttorneys.Attorneys[0].Email = email
			} else {
				donor.Attorneys.Attorneys[0].Email = email
			}
		}

		var attorneyID string
		if !isTrustCorporation {
			if isReplacement {
				attorneyID = donor.ReplacementAttorneys.Attorneys[0].ID
			} else {
				attorneyID = donor.Attorneys.Attorneys[0].ID
			}
		}

		donor.AttorneyDecisions = actor.AttorneyDecisions{How: actor.JointlyAndSeverally}
		donor.ReplacementAttorneyDecisions = actor.AttorneyDecisions{How: actor.JointlyAndSeverally}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		attorney, err := attorneyStore.Create(attorneyCtx, donorSessionID, attorneyID, isReplacement, isTrustCorporation)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			donor.SignedAt = time.Now()
			certificateProvider.Certificate = actor.Certificate{Agreed: donor.SignedAt.Add(time.Hour)}
		}

		if progress >= slices.Index(progressValues, "signedByAttorney") {
			attorney.Mobile = testMobile
			attorney.ContactLanguagePreference = localize.En
			attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
			attorney.Tasks.ReadTheLpa = actor.TaskCompleted
			attorney.Tasks.SignTheLpa = actor.TaskCompleted

			if isTrustCorporation {
				attorney.WouldLikeSecondSignatory = form.No
				attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
					FirstNames:        "A",
					LastName:          "Sign",
					ProfessionalTitle: "Assistant to the signer",
					LpaSignedAt:       donor.SignedAt,
					Confirmed:         donor.SignedAt.Add(2 * time.Hour),
				}}
			} else {
				attorney.LpaSignedAt = donor.SignedAt
				attorney.Confirmed = donor.SignedAt.Add(2 * time.Hour)
			}
		}

		if progress >= slices.Index(progressValues, "signedByAllAttorneys") {
			for isReplacement, list := range map[bool]actor.Attorneys{false: donor.Attorneys, true: donor.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donor.LpaID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, a.ID, isReplacement, false)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.LpaSignedAt = donor.SignedAt
					attorney.Confirmed = donor.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}

				if list.TrustCorporation.Name != "" {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donor.LpaID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, "", isReplacement, true)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						LpaSignedAt:       donor.SignedAt,
						Confirmed:         donor.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}
			}
		}

		if progress >= slices.Index(progressValues, "submitted") {
			donor.SubmittedAt = time.Now()
		}

		if progress == slices.Index(progressValues, "withdrawn") {
			donor.WithdrawnAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "registered") {
			donor.RegisteredAt = time.Now()
		}

		if err := donorStore.Put(donorCtx, donor); err != nil {
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
			shareCodeSender.UseTestCode()
		}

		if email != "" {
			shareCodeSender.SendAttorneys(donorCtx, page.AppData{
				SessionID: donorSessionID,
				LpaID:     donor.LpaID,
				Localizer: appData.Localizer,
			}, donor)

			http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
			return nil
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/attorney/" + donor.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
