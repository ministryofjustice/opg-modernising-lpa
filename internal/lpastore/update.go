package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

type updateRequest struct {
	Type    string                `json:"type"`
	Changes []updateRequestChange `json:"changes"`
}

type updateRequestChange struct {
	Key string `json:"key"`
	Old any    `json:"old"`
	New any    `json:"new"`
}

func (c *Client) sendUpdate(ctx context.Context, lpaUID string, actorUID actoruid.UID, body updateRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/lpas/"+lpaUID+"/updates", &buf)
	if err != nil {
		return err
	}

	return c.do(ctx, actorUID, req, nil)
}

func (c *Client) SendCertificateProvider(ctx context.Context, lpaUID string, certificateProvider *actor.CertificateProviderProvidedDetails, lpa *Lpa) error {
	body := updateRequest{
		Type: "CERTIFICATE_PROVIDER_SIGN",
		Changes: []updateRequestChange{
			{Key: "/certificateProvider/signedAt", New: certificateProvider.Certificate.Agreed},
			{Key: "/certificateProvider/contactLanguagePreference", New: certificateProvider.ContactLanguagePreference.String()},
		},
	}

	if certificateProvider.HomeAddress.Line1 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line1", New: certificateProvider.HomeAddress.Line1})
	}

	if certificateProvider.HomeAddress.Line2 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line2", New: certificateProvider.HomeAddress.Line2})
	}

	if certificateProvider.HomeAddress.Line3 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line3", New: certificateProvider.HomeAddress.Line3})
	}

	if certificateProvider.HomeAddress.TownOrCity != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/town", New: certificateProvider.HomeAddress.TownOrCity})
	}

	if certificateProvider.HomeAddress.Postcode != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/postcode", New: certificateProvider.HomeAddress.Postcode})
	}

	if certificateProvider.HomeAddress.Country != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/country", New: certificateProvider.HomeAddress.Country})
	}

	if lpa.CertificateProvider.Email != certificateProvider.Email {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/email", New: certificateProvider.Email, Old: lpa.CertificateProvider.Email})
	}

	if lpa.CertificateProvider.Channel != actor.ChannelOnline {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/channel", New: actor.ChannelOnline, Old: actor.ChannelPaper})
	}

	return c.sendUpdate(ctx, lpaUID, certificateProvider.UID, body)
}

func (c *Client) SendAttorney(ctx context.Context, donor *Lpa, attorney *actor.AttorneyProvidedDetails) error {
	var attorneyKey string
	if attorney.IsTrustCorporation && attorney.IsReplacement && donor.Attorneys.TrustCorporation.Name != "" {
		attorneyKey = "/trustCorporations/1"
	} else if attorney.IsTrustCorporation {
		attorneyKey = "/trustCorporations/0"
	} else if attorney.IsReplacement {
		attorneyKey = fmt.Sprintf("/attorneys/%d", len(donor.Attorneys.Attorneys)+donor.ReplacementAttorneys.Index(attorney.UID))
	} else {
		attorneyKey = fmt.Sprintf("/attorneys/%d", donor.Attorneys.Index(attorney.UID))
	}

	body := updateRequest{
		Type: "ATTORNEY_SIGN",
		Changes: []updateRequestChange{
			{Key: attorneyKey + "/mobile", New: attorney.Mobile},
			{Key: attorneyKey + "/contactLanguagePreference", New: attorney.ContactLanguagePreference.String()},
		},
	}

	if attorney.IsTrustCorporation {
		body.Type = "TRUST_CORPORATION_SIGN"

		body.Changes = append(body.Changes,
			updateRequestChange{Key: attorneyKey + "/signatories/0/firstNames", New: attorney.AuthorisedSignatories[0].FirstNames},
			updateRequestChange{Key: attorneyKey + "/signatories/0/lastName", New: attorney.AuthorisedSignatories[0].LastName},
			updateRequestChange{Key: attorneyKey + "/signatories/0/professionalTitle", New: attorney.AuthorisedSignatories[0].ProfessionalTitle},
			updateRequestChange{Key: attorneyKey + "/signatories/0/signedAt", New: attorney.AuthorisedSignatories[0].Confirmed},
		)

		if !attorney.AuthorisedSignatories[1].Confirmed.IsZero() {
			body.Changes = append(body.Changes,
				updateRequestChange{Key: attorneyKey + "/signatories/1/firstNames", New: attorney.AuthorisedSignatories[1].FirstNames},
				updateRequestChange{Key: attorneyKey + "/signatories/1/lastName", New: attorney.AuthorisedSignatories[1].LastName},
				updateRequestChange{Key: attorneyKey + "/signatories/1/professionalTitle", New: attorney.AuthorisedSignatories[1].ProfessionalTitle},
				updateRequestChange{Key: attorneyKey + "/signatories/1/signedAt", New: attorney.AuthorisedSignatories[1].Confirmed},
			)
		}
	} else {
		body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/signedAt", New: attorney.Confirmed})
	}

	return c.sendUpdate(ctx, donor.LpaUID, attorney.UID, body)
}
