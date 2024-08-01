package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

	resp, err := c.do(ctx, actorUID, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return responseError{
			name: fmt.Sprintf("expected 201 response but got %d", resp.StatusCode),
			body: string(body),
		}
	}

	return nil
}

func (c *Client) SendRegister(ctx context.Context, lpaUID string) error {
	body := updateRequest{
		Type: "REGISTER",
	}

	return c.sendUpdate(ctx, lpaUID, actoruid.Service, body)
}

func (c *Client) SendPerfect(ctx context.Context, lpaUID string) error {
	body := updateRequest{
		Type: "PERFECT",
	}

	return c.sendUpdate(ctx, lpaUID, actoruid.Service, body)
}

func (c *Client) SendCertificateProvider(ctx context.Context, certificateProvider *certificateproviderdata.Provided, lpa *Lpa) error {
	body := updateRequest{
		Type: "CERTIFICATE_PROVIDER_SIGN",
		Changes: []updateRequestChange{
			{Key: "/certificateProvider/signedAt", New: certificateProvider.SignedAt},
			{Key: "/certificateProvider/contactLanguagePreference", New: certificateProvider.ContactLanguagePreference.String()},
		},
	}

	if certificateProvider.HomeAddress.Line1 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line1", New: certificateProvider.HomeAddress.Line1, Old: lpa.CertificateProvider.Address.Line1})
	}

	if certificateProvider.HomeAddress.Line2 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line2", New: certificateProvider.HomeAddress.Line2, Old: lpa.CertificateProvider.Address.Line2})
	}

	if certificateProvider.HomeAddress.Line3 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line3", New: certificateProvider.HomeAddress.Line3, Old: lpa.CertificateProvider.Address.Line3})
	}

	if certificateProvider.HomeAddress.TownOrCity != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/town", New: certificateProvider.HomeAddress.TownOrCity, Old: lpa.CertificateProvider.Address.TownOrCity})
	}

	if certificateProvider.HomeAddress.Postcode != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/postcode", New: certificateProvider.HomeAddress.Postcode, Old: lpa.CertificateProvider.Address.Postcode})
	}

	if certificateProvider.HomeAddress.Country != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/country", New: certificateProvider.HomeAddress.Country, Old: lpa.CertificateProvider.Address.Country})
	}

	if certificateProvider.Email != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/email", New: certificateProvider.Email, Old: lpa.CertificateProvider.Email})
	}

	if lpa.CertificateProvider.Channel == donordata.ChannelPaper {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/channel", New: donordata.ChannelOnline, Old: donordata.ChannelPaper})
	}

	return c.sendUpdate(ctx, lpa.LpaUID, certificateProvider.UID, body)
}

func (c *Client) SendAttorney(ctx context.Context, lpa *Lpa, attorney *attorneydata.Provided) error {
	var attorneyKey string
	var lpaAttorney Attorney
	var lpaTrustCorp TrustCorporation

	if attorney.IsTrustCorporation && attorney.IsReplacement && lpa.Attorneys.TrustCorporation.Name != "" {
		attorneyKey = "/trustCorporations/1"
		lpaTrustCorp = lpa.ReplacementAttorneys.TrustCorporation
	} else if attorney.IsTrustCorporation {
		attorneyKey = "/trustCorporations/0"
		lpaTrustCorp = lpa.Attorneys.TrustCorporation
	} else if attorney.IsReplacement {
		attorneyKey = fmt.Sprintf("/attorneys/%d", len(lpa.Attorneys.Attorneys)+lpa.ReplacementAttorneys.Index(attorney.UID))
		lpaAttorney = lpa.ReplacementAttorneys.Attorneys[lpa.ReplacementAttorneys.Index(attorney.UID)]
	} else {
		attorneyKey = fmt.Sprintf("/attorneys/%d", lpa.Attorneys.Index(attorney.UID))
		lpaAttorney = lpa.Attorneys.Attorneys[lpa.Attorneys.Index(attorney.UID)]
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

		if lpaTrustCorp.Email != attorney.Email {
			body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/email", New: attorney.Email, Old: lpaTrustCorp.Email})
		}

		if lpaTrustCorp.Channel == donordata.ChannelPaper {
			body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/channel", New: donordata.ChannelOnline, Old: donordata.ChannelPaper})
		}

		body.Changes = append(body.Changes,
			updateRequestChange{Key: attorneyKey + "/signatories/0/firstNames", New: attorney.AuthorisedSignatories[0].FirstNames},
			updateRequestChange{Key: attorneyKey + "/signatories/0/lastName", New: attorney.AuthorisedSignatories[0].LastName},
			updateRequestChange{Key: attorneyKey + "/signatories/0/professionalTitle", New: attorney.AuthorisedSignatories[0].ProfessionalTitle},
			updateRequestChange{Key: attorneyKey + "/signatories/0/signedAt", New: attorney.AuthorisedSignatories[0].SignedAt},
		)

		if !attorney.AuthorisedSignatories[1].SignedAt.IsZero() {
			body.Changes = append(body.Changes,
				updateRequestChange{Key: attorneyKey + "/signatories/1/firstNames", New: attorney.AuthorisedSignatories[1].FirstNames},
				updateRequestChange{Key: attorneyKey + "/signatories/1/lastName", New: attorney.AuthorisedSignatories[1].LastName},
				updateRequestChange{Key: attorneyKey + "/signatories/1/professionalTitle", New: attorney.AuthorisedSignatories[1].ProfessionalTitle},
				updateRequestChange{Key: attorneyKey + "/signatories/1/signedAt", New: attorney.AuthorisedSignatories[1].SignedAt},
			)
		}
	} else {
		if attorney.Email != lpaAttorney.Email {
			body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/email", New: attorney.Email, Old: lpaAttorney.Email})
		}

		if lpaAttorney.Channel == donordata.ChannelPaper {
			body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/channel", New: donordata.ChannelOnline, Old: donordata.ChannelPaper})
		}

		body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/signedAt", New: attorney.SignedAt})
	}

	return c.sendUpdate(ctx, lpa.LpaUID, attorney.UID, body)
}

func (c *Client) SendCertificateProviderOptOut(ctx context.Context, lpaUID string, actorUID actoruid.UID) error {
	body := updateRequest{
		Type: "CERTIFICATE_PROVIDER_OPT_OUT",
	}

	return c.sendUpdate(ctx, lpaUID, actorUID, body)
}

func (c *Client) SendDonorConfirmIdentity(ctx context.Context, donor *donordata.Provided) error {
	body := updateRequest{
		Type: "DONOR_CONFIRM_IDENTITY",
		Changes: []updateRequestChange{
			{Key: "/donor/identityCheck/checkedAt", New: donor.DonorIdentityUserData.RetrievedAt, Old: nil},
			{Key: "/donor/identityCheck/type", New: "one-login", Old: nil},
		},
	}

	return c.sendUpdate(ctx, donor.LpaUID, donor.Donor.UID, body)
}

func (c *Client) SendCertificateProviderConfirmIdentity(ctx context.Context, lpaUID string, certificateProvider *certificateproviderdata.Provided) error {
	body := updateRequest{
		Type: "CERTIFICATE_PROVIDER_CONFIRM_IDENTITY",
		Changes: []updateRequestChange{
			{Key: "/certificateProvider/identityCheck/checkedAt", New: certificateProvider.IdentityUserData.RetrievedAt, Old: nil},
			{Key: "/certificateProvider/identityCheck/type", New: "one-login", Old: nil},
		},
	}

	return c.sendUpdate(ctx, lpaUID, certificateProvider.UID, body)
}
