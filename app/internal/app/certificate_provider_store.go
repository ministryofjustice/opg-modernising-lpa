package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type certificateProviderStore struct {
	dataStore DataStore
	now       func() time.Time
}

func (s *certificateProviderStore) Create(ctx context.Context) (*actor.CertificateProvider, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return &actor.CertificateProvider{}, err
	}

	cp := &actor.CertificateProvider{LpaID: data.LpaID}
	err = s.Put(ctx, cp)

	return cp, err
}

func (s *certificateProviderStore) Get(ctx context.Context) (*actor.CertificateProvider, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return &actor.CertificateProvider{}, err
	}

	if data.LpaID == "" {
		return &actor.CertificateProvider{}, errors.New("certificateProviderStore.Get requires LpaID to retrieve")
	}

	var certificateProvider actor.CertificateProvider

	pk := "LPA#" + data.LpaID

	if err := s.dataStore.GetOneByPartialSk(ctx, pk, "#CERTIFICATE_PROVIDER#", &certificateProvider); err != nil {
		return &actor.CertificateProvider{}, err
	}

	return &certificateProvider, nil
}

func (s *certificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProvider) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return errors.New("certificateProviderStore.Put requires LpaID and SessionID to retrieve")
	}

	pk := "LPA#" + data.LpaID
	sk := "#CERTIFICATE_PROVIDER#" + data.SessionID

	certificateProvider.UpdatedAt = s.now()

	return s.dataStore.Put(ctx, pk, sk, certificateProvider)
}
