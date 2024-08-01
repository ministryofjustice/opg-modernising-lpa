package fixtures

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

func CertificateProvider(
	tmpl template.Template,
	sessionStore *sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
	dynamoClient DynamoClient,
	organisationStore OrganisationStore,
	memberStore MemberStore,
	shareCodeStore ShareCodeStore,
) page.Handler {
	progressValues := []string{
		"paid",
		"signedByDonor",
		"confirmYourDetails",
		"confirmYourIdentity",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			lpaType                           = r.FormValue("lpa-type")
			progress                          = slices.Index(progressValues, r.FormValue("progress"))
			email                             = r.FormValue("email")
			donorEmail                        = r.FormValue("donorEmail")
			redirect                          = r.FormValue("redirect")
			asProfessionalCertificateProvider = r.FormValue("relationship") == "professional"
			certificateProviderSub            = r.FormValue("certificateProviderSub")
			shareCode                         = r.FormValue("withShareCode")
			useRealUID                        = r.FormValue("uid") == "real"
			donorChannel                      = r.FormValue("donorChannel")
			isSupported                       = r.FormValue("is-supported") == "1"
		)

		if certificateProviderSub == "" {
			certificateProviderSub = random.String(16)
		}

		if donorEmail == "" {
			donorEmail = testEmail
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: certificateProviderSub, DonorEmail: donorEmail})
		}

		var (
			donorSub                     = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
		)

		err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: certificateProviderSub, Email: testEmail})
		if err != nil {
			return err
		}

		var donorDetails *donordata.DonorProvidedDetails

		if donorChannel == "paper" {
			lpaID := random.UuidString()
			donorDetails = &donordata.DonorProvidedDetails{
				PK:                             dynamo.LpaKey(lpaID),
				SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:                          lpaID,
				LpaUID:                         random.UuidString(),
				CreatedAt:                      time.Now(),
				Version:                        1,
				HasSentApplicationUpdatedEvent: true,
				SignedAt:                       time.Now(),
			}

			if err := dynamoClient.Create(r.Context(), donorDetails); err != nil {
				return err
			}
		} else if isSupported {
			supporterCtx := page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: donorSessionID, Email: testEmail})

			member, err := memberStore.Create(supporterCtx, random.String(12), random.String(12))
			if err != nil {
				return err
			}

			org, err := organisationStore.Create(supporterCtx, member, random.String(12))
			if err != nil {
				return err
			}

			orgSession := &appcontext.SessionData{SessionID: donorSessionID, OrganisationID: org.ID}
			donorDetails, err = organisationStore.CreateLPA(page.ContextWithSessionData(r.Context(), orgSession))
			if err != nil {
				return err
			}

			if err := donorStore.Link(page.ContextWithSessionData(r.Context(), orgSession), actor.ShareCodeData{
				LpaKey:      donorDetails.PK,
				LpaOwnerKey: donorDetails.SK,
			}, donorDetails.Donor.Email); err != nil {
				return err
			}
		} else {
			donorDetails, err = donorStore.Create(page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &appcontext.SessionData{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})
		)

		donorDetails.Donor = makeDonor(donorEmail)

		donorDetails.Type = actor.LpaTypePropertyAndAffairs
		if lpaType == "personal-welfare" {
			donorDetails.Type = actor.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = donordata.CanBeUsedWhenCapacityLost
			donorDetails.LifeSustainingTreatmentOption = actor.LifeSustainingTreatmentOptionA
		} else {
			donorDetails.WhenCanTheLpaBeUsed = donordata.CanBeUsedWhenHasCapacity
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
		} else {
			donorDetails.LpaUID = makeUID()
		}

		donorDetails.Attorneys = donordata.Attorneys{
			Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
		}

		donorDetails.AttorneyDecisions = donordata.AttorneyDecisions{How: donordata.JointlyAndSeverally}

		donorDetails.CertificateProvider = makeCertificateProvider()
		if email != "" {
			donorDetails.CertificateProvider.Email = email
		}

		if asProfessionalCertificateProvider {
			donorDetails.CertificateProvider.Relationship = donordata.Professionally
		}

		certificateProvider, err := createCertificateProvider(certificateProviderCtx, shareCodeStore, certificateProviderStore, donorDetails.CertificateProvider.UID, donorDetails.SK, donorDetails.CertificateProvider.Email)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "paid") {
			donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, actor.Payment{
				PaymentReference: random.String(12),
				PaymentId:        random.String(12),
			})
			donorDetails.Tasks.PayForLpa = actor.PaymentTaskCompleted
		}

		if progress >= slices.Index(progressValues, "signedByDonor") {
			donorDetails.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskCompleted
			donorDetails.WitnessedByCertificateProviderAt = time.Now()
			donorDetails.SignedAt = time.Now()
		}

		if progress >= slices.Index(progressValues, "confirmYourDetails") {
			certificateProvider.DateOfBirth = date.New("1990", "1", "2")
			certificateProvider.ContactLanguagePreference = localize.En
			certificateProvider.Tasks.ConfirmYourDetails = actor.TaskCompleted

			if asProfessionalCertificateProvider {
				certificateProvider.HomeAddress = place.Address{
					Line1:      "6 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				}
			}
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			certificateProvider.IdentityUserData = identity.UserData{
				Status:      identity.StatusConfirmed,
				RetrievedAt: time.Now(),
				FirstNames:  donorDetails.CertificateProvider.FirstNames,
				LastName:    donorDetails.CertificateProvider.LastName,
				DateOfBirth: certificateProvider.DateOfBirth,
			}
			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted
		}

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}

		if !donorDetails.SignedAt.IsZero() && donorDetails.LpaUID != "" {
			if err := lpaStoreClient.SendLpa(donorCtx, donorDetails); err != nil {
				return err
			}
		}

		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if shareCode != "" {
			shareCodeSender.UseTestCode(shareCode)
		}

		if email != "" {
			shareCodeSender.SendCertificateProviderInvite(donorCtx, page.AppData{
				SessionID: donorSessionID,
				LpaID:     donorDetails.LpaID,
				Localizer: appData.Localizer,
			}, page.CertificateProviderInvite{
				LpaKey:                      donorDetails.PK,
				LpaOwnerKey:                 donorDetails.SK,
				LpaUID:                      donorDetails.LpaUID,
				Type:                        donorDetails.Type,
				DonorFirstNames:             donorDetails.Donor.FirstNames,
				DonorFullName:               donorDetails.Donor.FullName(),
				CertificateProviderUID:      donorDetails.CertificateProvider.UID,
				CertificateProviderFullName: donorDetails.CertificateProvider.FullName(),
				CertificateProviderEmail:    donorDetails.CertificateProvider.Email,
			})
		}

		switch redirect {
		case "":
			redirect = page.Paths.Dashboard.Format()
		case page.Paths.CertificateProviderStart.Format():
			redirect = page.Paths.CertificateProviderStart.Format()
		case page.Paths.CertificateProvider.EnterReferenceNumberOptOut.Format():
			redirect = page.Paths.CertificateProvider.EnterReferenceNumberOptOut.Format()
		default:
			redirect = "/certificate-provider/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
