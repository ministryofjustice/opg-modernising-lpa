package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DonorStore interface {
	Create(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Link(ctx context.Context, shareCode actor.ShareCodeData) error
	Put(ctx context.Context, donorProvidedDetails *actor.DonorProvidedDetails) error
}

type CertificateProviderStore interface {
	Create(ctx context.Context, shareCode actor.ShareCodeData, email string) (*actor.CertificateProviderProvidedDetails, error)
	Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error
}

type AttorneyStore interface {
	Create(ctx context.Context, shareCode actor.ShareCodeData, email string) (*actor.AttorneyProvidedDetails, error)
	Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error
}

func Attorney(
	tmpl template.Template,
	sessionStore *sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
	organisationStore OrganisationStore,
	memberStore MemberStore,
	shareCodeStore ShareCodeStore,
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
		acceptCookiesConsent(w)

		var (
			isReplacement      = r.FormValue("is-replacement") == "1"
			isTrustCorporation = r.FormValue("is-trust-corporation") == "1"
			isSupported        = r.FormValue("is-supported") == "1"
			lpaType            = r.FormValue("lpa-type")
			progress           = slices.Index(progressValues, r.FormValue("progress"))
			email              = r.FormValue("email")
			redirect           = r.FormValue("redirect")
			attorneySub        = r.FormValue("attorneySub")
			shareCode          = r.FormValue("withShareCode")
			useRealUID         = r.FormValue("uid") == "real"
		)

		if attorneySub == "" {
			attorneySub = random.String(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: attorneySub})
		}

		if lpaType == "personal-welfare" && isTrustCorporation {
			return tmpl(w, &fixturesData{App: appData, Errors: validation.With("", validation.CustomError{Label: "Can't add a trust corporation to a personal welfare LPA"})})
		}

		var (
			donorSub                     = random.String(16)
			certificateProviderSub       = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(attorneySub))
		)

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: attorneySub, Email: testEmail}); err != nil {
			return err
		}

		createFn := donorStore.Create
		createSession := &page.SessionData{SessionID: donorSessionID}
		if isSupported {
			createFn = organisationStore.CreateLPA

			supporterCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, Email: testEmail})

			member, err := memberStore.Create(supporterCtx, random.String(12), random.String(12))
			if err != nil {
				return err
			}

			org, err := organisationStore.Create(supporterCtx, member, random.String(12))
			if err != nil {
				return err
			}

			createSession.OrganisationID = org.ID
		}
		donorDetails, err := createFn(page.ContextWithSessionData(r.Context(), createSession))
		if err != nil {
			return err
		}

		if isSupported {
			if err := donorStore.Link(page.ContextWithSessionData(r.Context(), createSession), actor.ShareCodeData{
				LpaKey:      donorDetails.PK,
				LpaOwnerKey: donorDetails.SK,
			}); err != nil {
				return err
			}
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})
			attorneyCtx            = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: attorneySessionID, LpaID: donorDetails.LpaID})
		)

		donorDetails.SignedAt = time.Now()
		donorDetails.Donor = makeDonor()

		if lpaType == "personal-welfare" && !isTrustCorporation {
			donorDetails.Type = actor.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
			donorDetails.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
		} else {
			donorDetails.Type = actor.LpaTypePropertyAndAffairs
			donorDetails.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenHasCapacity
		}

		if useRealUID {
			if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
				LpaID:          donorDetails.LpaID,
				DonorSessionID: donorSessionID,
				Type:           donorDetails.Type.String(),
				Donor: uid.DonorDetails{
					Name:     donorDetails.Donor.FullName(),
					Dob:      donorDetails.Donor.DateOfBirth,
					Postcode: donorDetails.Donor.Address.Postcode,
				},
			}); err != nil {
				return err
			}

			donorDetails.HasSentUidRequestedEvent = true
		} else {
			donorDetails.LpaUID = makeUID()
		}

		donorDetails.CertificateProvider = makeCertificateProvider()

		donorDetails.Attorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(attorneyNames[0])},
			TrustCorporation: makeTrustCorporation("First Choice Trust Corporation Ltd."),
		}
		donorDetails.ReplacementAttorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(replacementAttorneyNames[0])},
			TrustCorporation: makeTrustCorporation("Second Choice Trust Corporation Ltd."),
		}

		if email != "" {
			if isTrustCorporation && isReplacement {
				donorDetails.ReplacementAttorneys.TrustCorporation.Email = email
			} else if isTrustCorporation {
				donorDetails.Attorneys.TrustCorporation.Email = email
			} else if isReplacement {
				donorDetails.ReplacementAttorneys.Attorneys[0].Email = email
			} else {
				donorDetails.Attorneys.Attorneys[0].Email = email
			}
		}

		var attorneyUID actoruid.UID
		if isTrustCorporation && isReplacement {
			attorneyUID = donorDetails.ReplacementAttorneys.TrustCorporation.UID
		} else if isTrustCorporation {
			attorneyUID = donorDetails.Attorneys.TrustCorporation.UID
		} else if isReplacement {
			attorneyUID = donorDetails.ReplacementAttorneys.Attorneys[0].UID
		} else {
			attorneyUID = donorDetails.Attorneys.Attorneys[0].UID
		}

		donorDetails.AttorneyDecisions = actor.AttorneyDecisions{How: actor.JointlyAndSeverally}
		donorDetails.ReplacementAttorneyDecisions = actor.AttorneyDecisions{How: actor.JointlyAndSeverally}
		donorDetails.HowShouldReplacementAttorneysStepIn = actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct

		certificateProvider, err := createCertificateProvider(certificateProviderCtx, shareCodeStore, certificateProviderStore, donorDetails.CertificateProvider.UID, donorDetails.SK, donorDetails.CertificateProvider.Email)
		if err != nil {
			return err
		}

		certificateProvider.ContactLanguagePreference = localize.En

		attorney, err := createAttorney(
			attorneyCtx,
			shareCodeStore,
			attorneyStore,
			attorneyUID,
			isReplacement,
			isTrustCorporation,
			donorDetails.SK,
			donorDetails.CertificateProvider.Email,
		)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			donorDetails.SignedAt = time.Now()
			certificateProvider.Certificate = actor.Certificate{Agreed: donorDetails.SignedAt.Add(time.Hour)}
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
					Confirmed:         donorDetails.SignedAt.Add(2 * time.Hour),
				}}
			} else {
				attorney.Confirmed = donorDetails.SignedAt.Add(2 * time.Hour)
			}
		}

		var signings []*actor.AttorneyProvidedDetails
		if progress >= slices.Index(progressValues, "signedByAllAttorneys") {
			for isReplacement, list := range map[bool]actor.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

					attorney, err := createAttorney(
						ctx,
						shareCodeStore,
						attorneyStore,
						a.UID,
						isReplacement,
						false,
						donorDetails.SK,
						a.Email,
					)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.ContactLanguagePreference = localize.En
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.Confirmed = donorDetails.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}

					signings = append(signings, attorney)
				}

				if list.TrustCorporation.Name != "" {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: donorDetails.LpaID})

					attorney, err := createAttorney(
						ctx,
						shareCodeStore,
						attorneyStore,
						list.TrustCorporation.UID,
						isReplacement,
						true,
						donorDetails.SK,
						list.TrustCorporation.Email,
					)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.ContactLanguagePreference = localize.En
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						Confirmed:         donorDetails.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}

					signings = append(signings, attorney)
				}
			}
		}

		if progress >= slices.Index(progressValues, "submitted") {
			donorDetails.SubmittedAt = time.Now()
		}

		if progress == slices.Index(progressValues, "withdrawn") {
			donorDetails.WithdrawnAt = time.Now()
		}

		registered := false
		if progress >= slices.Index(progressValues, "registered") {
			registered = true
		}

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}
		if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
			return err
		}

		if donorDetails.LpaUID != "" {
			if err := lpaStoreClient.SendLpa(donorCtx, donorDetails); err != nil {
				return fmt.Errorf("problem sending lpa: %w", err)
			}

			lpa, err := lpaStoreClient.Lpa(donorCtx, donorDetails.LpaUID)
			if err != nil {
				return fmt.Errorf("problem getting lpa: %w", err)
			}

			if err := lpaStoreClient.SendCertificateProvider(donorCtx, certificateProvider, lpa); err != nil {
				return fmt.Errorf("problem sending certificate provider: %w", err)
			}

			for _, attorney := range signings {
				if err := lpaStoreClient.SendAttorney(donorCtx, lpa, attorney); err != nil {
					return fmt.Errorf("problem sending attorney: %w", err)
				}
			}

			if registered {
				if err := lpaStoreClient.SendRegister(donorCtx, donorDetails.LpaUID); err != nil {
					return fmt.Errorf("problem sending register: %w", err)
				}
			}
		}

		// should only be used in tests as otherwise people can read their emails...
		if shareCode != "" {
			shareCodeSender.UseTestCode(shareCode)
		}

		if email != "" {
			lpa, err := lpaStoreClient.Lpa(r.Context(), donorDetails.LpaUID)
			if err != nil {
				return err
			}

			lpa.LpaKey = donorDetails.PK
			lpa.LpaOwnerKey = donorDetails.SK

			shareCodeSender.SendAttorneys(donorCtx, page.AppData{
				SessionID: donorSessionID,
				LpaID:     donorDetails.LpaID,
				Localizer: appData.Localizer,
			}, lpa)

			http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
			return nil
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/attorney/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
