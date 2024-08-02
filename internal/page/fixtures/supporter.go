package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type OrganisationStore interface {
	Create(context.Context, *actor.Member, string) (*actor.Organisation, error)
	CreateLPA(context.Context) (*donordata.Provided, error)
}

type MemberStore interface {
	Create(ctx context.Context, firstNames, lastName string) (*actor.Member, error)
	CreateFromInvite(ctx context.Context, invite *actor.MemberInvite) error
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error
	Put(ctx context.Context, member *actor.Member) error
}

type ShareCodeStore interface {
	Put(ctx context.Context, actorType actor.Type, shareCode string, data sharecode.Data) error
	PutDonor(ctx context.Context, code string, data sharecode.Data) error
}

func Supporter(
	sessionStore *sesh.Store,
	organisationStore OrganisationStore,
	donorStore DonorStore,
	memberStore MemberStore,
	dynamoClient DynamoClient,
	searchClient *search.Client,
	shareCodeStore ShareCodeStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	documentStore DocumentStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			invitedMembers = r.FormValue("invitedMembers")
			lpa            = r.FormValue("lpa")
			members        = r.FormValue("members")
			organisation   = r.FormValue("organisation")
			redirect       = r.FormValue("redirect")
			asMember       = r.FormValue("asMember")
			permission     = r.FormValue("permission")
			expireInvites  = r.FormValue("expireInvites") == "1"
			suspended      = r.FormValue("suspended") == "1"
			setLPAProgress = r.FormValue("setLPAProgress") == "1"
			accessCode     = r.FormValue("accessCode")
			linkDonor      = r.FormValue("linkDonor") == "1"

			supporterSub       = random.String(16)
			supporterSessionID = base64.StdEncoding.EncodeToString([]byte(supporterSub))
			supporterCtx       = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: supporterSessionID, Email: testEmail})
		)

		loginSession := &sesh.LoginSession{Sub: supporterSub, Email: testEmail}

		if organisation == "1" {
			member, err := memberStore.Create(supporterCtx, random.String(12), random.String(12))
			if err != nil {
				return err
			}

			org, err := organisationStore.Create(supporterCtx, member, random.String(12))
			if err != nil {
				return err
			}

			loginSession.OrganisationID = org.ID
			loginSession.OrganisationName = org.Name

			organisationCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{OrganisationID: org.ID})

			if suspended {
				member.Status = actor.StatusSuspended

				if err := memberStore.Put(organisationCtx, member); err != nil {
					return err
				}
			}

			if accessCode != "" {
				donor, err := organisationStore.CreateLPA(organisationCtx)
				if err != nil {
					return err
				}
				donorCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{OrganisationID: org.ID, LpaID: donor.LpaID, SessionID: random.String(12)})

				donor.LpaUID = makeUID()
				donor.Donor = makeDonor(testEmail)
				donor.Type = donordata.LpaTypePropertyAndAffairs
				donor.CertificateProvider = makeCertificateProvider()
				donor.Attorneys = donordata.Attorneys{
					Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0])},
				}
				donor.Tasks.YourDetails = task.StateCompleted
				donor.Tasks.ChooseAttorneys = task.StateCompleted
				donor.Tasks.CertificateProvider = task.StateCompleted

				if err := donorStore.Put(donorCtx, donor); err != nil {
					return err
				}

				shareCodeData := sharecode.Data{
					LpaOwnerKey:  dynamo.LpaOwnerKey(org.PK),
					LpaKey:       donor.PK,
					ActorUID:     donor.Donor.UID,
					InviteSentTo: "email@example.com",
				}

				if err != nil {
					return err
				}

				if err := shareCodeStore.PutDonor(r.Context(), accessCode, shareCodeData); err != nil {
					return err
				}

				if linkDonor {
					shareCodeData.PK = dynamo.ShareKey(dynamo.DonorShareKey(accessCode))
					shareCodeData.SK = dynamo.ShareSortKey(dynamo.DonorInviteKey(org.PK, shareCodeData.LpaKey))
					shareCodeData.UpdatedAt = time.Now()

					if err := donorStore.Link(donorCtx, shareCodeData, donor.Donor.Email); err != nil {
						return err
					}

					waitForLPAIndex(searchClient, organisationCtx)
				}
			}

			if lpaCount, err := strconv.Atoi(lpa); err == nil {
				donorFixtureData := setFixtureData(r)

				for range lpaCount {
					donor, err := organisationStore.CreateLPA(organisationCtx)
					if err != nil {
						return err
					}
					donorCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{OrganisationID: org.ID, LpaID: donor.LpaID})

					donor.LpaUID = makeUID()
					donor.Donor = makeDonor(testEmail)
					donor.Type = donordata.LpaTypePropertyAndAffairs
					donor.CertificateProvider = makeCertificateProvider()
					donor.Attorneys = donordata.Attorneys{
						Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0])},
					}

					var fns []func(context.Context, *lpastore.Client, *lpastore.Lpa) error
					if setLPAProgress {
						donor, fns, err = updateLPAProgress(donorFixtureData, donor, random.String(16), r, certificateProviderStore, attorneyStore, documentStore, eventClient, shareCodeStore)
						if err != nil {
							return err
						}
					}

					if err := donorStore.Put(donorCtx, donor); err != nil {
						return err
					}
					if !donor.SignedAt.IsZero() && donor.LpaUID != "" {
						if err := lpaStoreClient.SendLpa(donorCtx, donor); err != nil {
							return err
						}

						lpa, err := lpaStoreClient.Lpa(donorCtx, donor.LpaUID)
						if err != nil {
							return fmt.Errorf("problem getting lpa: %w", err)
						}

						for _, fn := range fns {
							if err := fn(donorCtx, lpaStoreClient, lpa); err != nil {
								return err
							}
						}
					}
				}

				waitForLPAIndex(searchClient, organisationCtx)
			}

			if invitedMembers != "" {
				n, err := strconv.Atoi(invitedMembers)
				if err != nil {
					return fmt.Errorf("invitedMembers should be a number")
				}

				for i, member := range invitedOrgMemberNames {
					if i == n {
						break
					}

					email := strings.ToLower(fmt.Sprintf("%s-%s@example.org", member.Firstnames, member.Lastname))

					now := time.Now()
					if expireInvites {
						now = now.Add(time.Hour * -time.Duration(48))
					}

					invite := &actor.MemberInvite{
						PK:               dynamo.OrganisationKey(org.ID),
						SK:               dynamo.MemberInviteKey(email),
						CreatedAt:        now,
						OrganisationID:   org.ID,
						OrganisationName: org.Name,
						Email:            email,
						FirstNames:       member.Firstnames,
						LastName:         member.Lastname,
						Permission:       actor.PermissionAdmin,
						ReferenceNumber:  random.String(12),
					}

					if err := dynamoClient.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{OrganisationID: org.ID}), invite); err != nil {
						return fmt.Errorf("error creating member invite: %w", err)
					}
				}
			}

			if members != "" {
				n, err := strconv.Atoi(members)
				if err != nil {
					return fmt.Errorf("members should be a number")
				}

				memberEmailSub := make(map[string]string)

				permission, err := actor.ParsePermission(permission)
				if err != nil {
					permission = actor.PermissionNone
				}

				for i, member := range orgMemberNames {
					if i == n {
						break
					}

					email := strings.ToLower(fmt.Sprintf("%s-%s@example.org", member.Firstnames, member.Lastname))
					sub := []byte(random.String(16))
					memberCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: base64.StdEncoding.EncodeToString(sub), Email: email})

					if err = memberStore.CreateFromInvite(
						memberCtx,
						&actor.MemberInvite{
							PK:              dynamo.OrganisationKey(random.String(12)),
							SK:              dynamo.MemberInviteKey(random.String(12)),
							CreatedAt:       time.Now(),
							UpdatedAt:       time.Now(),
							OrganisationID:  org.ID,
							Email:           email,
							FirstNames:      member.Firstnames,
							LastName:        member.Lastname,
							Permission:      permission,
							ReferenceNumber: random.String(12),
						},
					); err != nil {
						return err
					}

					memberEmailSub[email] = string(sub)
				}

				if sub, found := memberEmailSub[asMember]; found {
					loginSession.Email = asMember
					loginSession.Sub = sub
				}
			}
		}

		if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
			return err
		}

		if redirect != page.Paths.Supporter.EnterOrganisationName.Format() && redirect != page.Paths.Supporter.EnterYourName.Format() && redirect != page.Paths.EnterAccessCode.Format() {
			redirect = "/supporter" + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}

func waitForLPAIndex(searchClient *search.Client, organisationCtx context.Context) {
	count := 0

	for range time.Tick(time.Second) {
		resp, err := searchClient.Query(organisationCtx, search.QueryRequest{
			Page:     1,
			PageSize: 1,
		})
		if err != nil {
			log.Println("error waiting for LPA Index:", err)
		}

		if count > 10 {
			return
		}
		count++

		if resp != nil && len(resp.Keys) > 0 {
			break
		}
	}
}
