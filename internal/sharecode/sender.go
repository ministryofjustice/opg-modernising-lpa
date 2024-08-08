package sharecode

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]any) string
	FormatDate(t date.TimeOrDate) string
	FormatTime(t time.Time) string
	FormatDateTime(t time.Time) string
	Possessive(s string) string
	SetShowTranslationKeys(s bool)
	ShowTranslationKeys() bool
	T(messageID string) string
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (sharecodedata.Data, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data sharecodedata.Data) error
}

type NotifyClient interface {
	SendActorEmail(context context.Context, to, lpaUID string, email notify.Email) error
	SendActorSMS(context context.Context, to, lpaUID string, sms notify.SMS) error
}

type EventClient interface {
	SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error
	SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error
}

type Sender struct {
	testCode       string
	shareCodeStore ShareCodeStore
	notifyClient   NotifyClient
	appPublicURL   string
	randomString   func(int) string
	eventClient    EventClient
}

func NewSender(shareCodeStore ShareCodeStore, notifyClient NotifyClient, appPublicURL string, randomString func(int) string, eventClient EventClient) *Sender {
	return &Sender{
		shareCodeStore: shareCodeStore,
		notifyClient:   notifyClient,
		appPublicURL:   appPublicURL,
		randomString:   randomString,
		eventClient:    eventClient,
	}
}

func (s *Sender) UseTestCode(shareCode string) {
	s.testCode = shareCode
}

type CertificateProviderInvite struct {
	LpaKey                      dynamo.LpaKeyType
	LpaOwnerKey                 dynamo.LpaOwnerKeyType
	LpaUID                      string
	Type                        lpadata.LpaType
	DonorFirstNames             string
	DonorFullName               string
	CertificateProviderUID      actoruid.UID
	CertificateProviderFullName string
	CertificateProviderEmail    string
}

func (s *Sender) SendCertificateProviderInvite(ctx context.Context, appData appcontext.Data, invite CertificateProviderInvite) error {
	shareCode, err := s.createShareCode(ctx, invite.LpaKey, invite.LpaOwnerKey, invite.CertificateProviderUID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	return s.sendEmail(ctx, invite.CertificateProviderEmail, invite.LpaUID, notify.CertificateProviderInviteEmail{
		CertificateProviderFullName:  invite.CertificateProviderFullName,
		DonorFullName:                invite.DonorFullName,
		LpaType:                      localize.LowerFirst(appData.Localizer.T(invite.Type.String())),
		CertificateProviderStartURL:  fmt.Sprintf("%s%s", s.appPublicURL, page.PathCertificateProviderStart),
		DonorFirstNames:              invite.DonorFirstNames,
		DonorFirstNamesPossessive:    appData.Localizer.Possessive(invite.DonorFirstNames),
		WhatLpaCovers:                appData.Localizer.T(invite.Type.WhatLPACoversTransKey()),
		ShareCode:                    shareCode,
		CertificateProviderOptOutURL: fmt.Sprintf("%s%s", s.appPublicURL, page.PathCertificateProviderEnterReferenceNumberOptOut),
	})
}

func (s *Sender) SendCertificateProviderPrompt(ctx context.Context, appData appcontext.Data, donor *donordata.Provided) error {
	shareCode, err := s.createShareCode(ctx, donor.PK, donor.SK, donor.CertificateProvider.UID, actor.TypeCertificateProvider)
	if err != nil {
		return err
	}

	if donor.CertificateProvider.CarryOutBy.IsPaper() {
		return s.sendPaperForm(ctx, donor.LpaUID, actor.TypeCertificateProvider, donor.CertificateProvider.UID, shareCode)
	}

	return s.sendEmail(ctx, donor.CertificateProvider.Email, donor.LpaUID, notify.CertificateProviderProvideCertificatePromptEmail{
		CertificateProviderFullName: donor.CertificateProvider.FullName(),
		DonorFullName:               donor.Donor.FullName(),
		LpaType:                     localize.LowerFirst(appData.Localizer.T(donor.Type.String())),
		CertificateProviderStartURL: fmt.Sprintf("%s%s", s.appPublicURL, page.PathCertificateProviderStart),
		ShareCode:                   shareCode,
	})
}

func (s *Sender) SendAttorneys(ctx context.Context, appData appcontext.Data, donor *lpadata.Lpa) error {
	if err := s.sendTrustCorporation(ctx, appData, donor, donor.Attorneys.TrustCorporation); err != nil {
		return err
	}
	if err := s.sendReplacementTrustCorporation(ctx, appData, donor, donor.ReplacementAttorneys.TrustCorporation); err != nil {
		return err
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if err := s.sendOriginalAttorney(ctx, appData, donor, attorney); err != nil {
			return err
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if err := s.sendReplacementAttorney(ctx, appData, donor, attorney); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) sendOriginalAttorney(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, attorney lpadata.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, attorney.UID, actor.TypeAttorney)
	if err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, attorney.Email, lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.appPublicURL + page.PathAttorneyStart.Format(),
			ShareCode:                 shareCode,
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendReplacementAttorney(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, attorney lpadata.Attorney) error {
	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, attorney.UID, actor.TypeReplacementAttorney)
	if err != nil {
		return err
	}

	if attorney.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementAttorney, attorney.UID, shareCode)
	}

	return s.sendEmail(ctx, attorney.Email, lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          attorney.FullName(),
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      s.appPublicURL + page.PathAttorneyStart.Format(),
			ShareCode:                 shareCode,
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendTrustCorporation(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, trustCorporation lpadata.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, trustCorporation.UID, actor.TypeTrustCorporation)
	if err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, trustCorporation.Email, lpa.LpaUID,
		notify.InitialOriginalAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, page.PathAttorneyStart),
			ShareCode:                 shareCode,
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) sendReplacementTrustCorporation(ctx context.Context, appData appcontext.Data, lpa *lpadata.Lpa, trustCorporation lpadata.TrustCorporation) error {
	if trustCorporation.Name == "" {
		return nil
	}

	shareCode, err := s.createShareCode(ctx, lpa.LpaKey, lpa.LpaOwnerKey, trustCorporation.UID, actor.TypeReplacementTrustCorporation)
	if err != nil {
		return err
	}

	if trustCorporation.Email == "" {
		return s.sendPaperForm(ctx, lpa.LpaUID, actor.TypeReplacementTrustCorporation, trustCorporation.UID, shareCode)
	}

	return s.sendEmail(ctx, trustCorporation.Email, lpa.LpaUID,
		notify.InitialReplacementAttorneyEmail{
			AttorneyFullName:          trustCorporation.Name,
			DonorFirstNames:           lpa.Donor.FirstNames,
			DonorFirstNamesPossessive: appData.Localizer.Possessive(lpa.Donor.FirstNames),
			DonorFullName:             lpa.Donor.FullName(),
			LpaType:                   localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
			AttorneyStartPageURL:      fmt.Sprintf("%s%s", s.appPublicURL, page.PathAttorneyStart),
			ShareCode:                 shareCode,
			AttorneyOptOutURL:         s.appPublicURL + page.PathAttorneyEnterReferenceNumberOptOut.Format(),
		})
}

func (s *Sender) createShareCode(ctx context.Context, lpaKey dynamo.LpaKeyType, lpaOwnerKey dynamo.LpaOwnerKeyType, actorUID actoruid.UID, actorType actor.Type) (string, error) {
	shareCode := s.randomString(12)
	if s.testCode != "" {
		shareCode = s.testCode
		s.testCode = ""
	}

	shareCodeData := sharecodedata.Data{
		LpaKey:                lpaKey,
		LpaOwnerKey:           lpaOwnerKey,
		ActorUID:              actorUID,
		IsReplacementAttorney: actorType == actor.TypeReplacementAttorney || actorType == actor.TypeReplacementTrustCorporation,
		IsTrustCorporation:    actorType == actor.TypeTrustCorporation || actorType == actor.TypeReplacementTrustCorporation,
	}

	if err := s.shareCodeStore.Put(ctx, actorType, shareCode, shareCodeData); err != nil {
		return "", fmt.Errorf("creating share failed: %w", err)
	}

	return shareCode, nil
}

func (s *Sender) sendEmail(ctx context.Context, to string, lpaUID string, email notify.Email) error {
	if err := s.notifyClient.SendActorEmail(ctx, to, lpaUID, email); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}

	return nil
}

func (s *Sender) sendPaperForm(ctx context.Context, lpaUID string, actorType actor.Type, actorUID actoruid.UID, shareCode string) error {
	return s.eventClient.SendPaperFormRequested(ctx, event.PaperFormRequested{
		UID:        lpaUID,
		ActorType:  actorType.String(),
		ActorUID:   actorUID,
		AccessCode: shareCode,
	})
}
