package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OrganisationStore interface {
	Create(context.Context, *actor.Member, string) (*actor.Organisation, error)
	CreateLPA(context.Context) (*actor.DonorProvidedDetails, error)
}

type MemberStore interface {
	Create(ctx context.Context, firstNames, lastName string) (*actor.Member, error)
	CreateFromInvite(ctx context.Context, invite *actor.MemberInvite) error
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error
	Put(ctx context.Context, member *actor.Member) error
}

type ShareCodeStore interface {
	Linked(ctx context.Context, data actor.ShareCodeData, email string) error
	PutDonor(ctx context.Context, code string, data actor.ShareCodeData) error
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
) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
			supporterCtx       = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: supporterSessionID, Email: testEmail})
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

			organisationCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID})

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
				donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID, LpaID: donor.LpaID, SessionID: random.String(12)})

				donor.LpaUID = makeUID()
				donor.Donor = makeDonor()
				donor.Type = actor.LpaTypePropertyAndAffairs

				donor.Attorneys = actor.Attorneys{
					Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
				}

				if err := donorStore.Put(donorCtx, donor); err != nil {
					return err
				}

				shareCodeData := actor.ShareCodeData{
					SessionID:    org.ID,
					LpaID:        donor.LpaID,
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
					shareCodeData.PK = "DONORSHARE#" + accessCode
					shareCodeData.SK = "DONORINVITE#" + shareCodeData.SessionID + "#" + shareCodeData.LpaID
					shareCodeData.UpdatedAt = time.Now()

					if err := donorStore.Link(donorCtx, shareCodeData); err != nil {
						return err
					}

					if err := shareCodeStore.Linked(donorCtx, shareCodeData, shareCodeData.InviteSentTo); err != nil {
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
					donorCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID, LpaID: donor.LpaID})

					donor.LpaUID = makeUID()
					donor.Donor = makeDonor()
					donor.Type = actor.LpaTypePropertyAndAffairs

					donor.Attorneys = actor.Attorneys{
						Attorneys: []actor.Attorney{makeAttorney(attorneyNames[0])},
					}

					if setLPAProgress {
						donor, err = updateLPAProgress(donorFixtureData, donor, random.String(16), r, certificateProviderStore, attorneyStore, documentStore, eventClient)
						if err != nil {
							return err
						}
					}

					if err := donorStore.Put(donorCtx, donor); err != nil {
						return err
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
						PK:               "ORGANISATION#" + org.ID,
						SK:               "MEMBERINVITE#" + base64.StdEncoding.EncodeToString([]byte(email)),
						CreatedAt:        now,
						OrganisationID:   org.ID,
						OrganisationName: org.Name,
						Email:            email,
						FirstNames:       member.Firstnames,
						LastName:         member.Lastname,
						Permission:       actor.PermissionAdmin,
						ReferenceNumber:  random.String(12),
					}

					if err := dynamoClient.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID}), invite); err != nil {
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
					memberCtx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: base64.StdEncoding.EncodeToString(sub), Email: email})

					if err = memberStore.CreateFromInvite(
						memberCtx,
						&actor.MemberInvite{
							PK:              random.String(12),
							SK:              random.String(12),
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
	for range time.Tick(time.Second) {
		if resp, _ := searchClient.Query(organisationCtx, search.QueryRequest{
			Page:     1,
			PageSize: 1,
		}); resp != nil && len(resp.Keys) > 0 {
			break
		}
	}
}
