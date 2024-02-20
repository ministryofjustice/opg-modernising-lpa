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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OrganisationStore interface {
	Create(context.Context, string) (*actor.Organisation, error)
	CreateLPA(context.Context) (*actor.DonorProvidedDetails, error)
}

type MemberStore interface {
	CreateMember(ctx context.Context, invite *actor.MemberInvite) error
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error
}

func Supporter(sessionStore sesh.Store, organisationStore OrganisationStore, donorStore DonorStore, memberStore MemberStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			invitedMembers = r.FormValue("invitedMembers")
			lpa            = r.FormValue("lpa")
			members        = r.FormValue("members")
			organisation   = r.FormValue("organisation")
			redirect       = r.FormValue("redirect")

			supporterSub       = random.String(16)
			supporterSessionID = base64.StdEncoding.EncodeToString([]byte(supporterSub))
			ctx                = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: supporterSessionID, Email: testEmail})
		)

		loginSession := &sesh.LoginSession{Sub: supporterSub, Email: testEmail}

		if organisation == "1" {
			org, err := organisationStore.Create(ctx, random.String(12))
			if err != nil {
				return err
			}

			loginSession.OrganisationID = org.ID
			loginSession.OrganisationName = org.Name

			if lpa == "1" {
				donor, err := organisationStore.CreateLPA(page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID}))
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

				if err := donorStore.Put(donorCtx, donor); err != nil {
					return err
				}
			}

			if invitedMembers != "" {
				n, err := strconv.Atoi(invitedMembers)

				for i, member := range invitedOrgMemberNames {
					if i == n {
						break
					}

					if err = memberStore.CreateMemberInvite(page.ContextWithSessionData(r.Context(), &page.SessionData{OrganisationID: org.ID}), org, member.Firstnames, member.Lastname, strings.ToLower(fmt.Sprintf("%s-%s@example.org", member.Firstnames, member.Lastname)), random.String(12), actor.Admin); err != nil {
						return err
					}
				}
			}

			if members != "" {
				n, err := strconv.Atoi(members)

				for i, member := range orgMemberNames {
					if i == n {
						break
					}

					if err = memberStore.CreateMember(
						page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(12)}),
						&actor.MemberInvite{
							PK:              random.String(12),
							SK:              random.String(12),
							CreatedAt:       time.Now(),
							UpdatedAt:       time.Now(),
							OrganisationID:  org.ID,
							Email:           strings.ToLower(fmt.Sprintf("%s-%s@example.org", member.Firstnames, member.Lastname)),
							FirstNames:      member.Firstnames,
							LastName:        member.Lastname,
							Permission:      actor.Admin,
							ReferenceNumber: random.String(12),
						},
					); err != nil {
						return err
					}
				}
			}
		}

		if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
			return err
		}

		if redirect != page.Paths.Supporter.EnterOrganisationName.Format() {
			redirect = "/supporter/" + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
