package donordata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/names"
)

type Voucher struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Email      string
	Allowed    bool
	FailedAt   time.Time
}

func (v Voucher) FullName() string {
	return v.FirstNames + " " + v.LastName
}

func (v Voucher) Matches(donor *Provided) (match []actor.Type) {
	if v.FirstNames == "" && v.LastName == "" {
		return nil
	}

	if names.EqualFull(donor.Donor, v) {
		match = append(match, actor.TypeDonor)
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if names.EqualFull(attorney, v) {
			match = append(match, actor.TypeAttorney)
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if names.EqualFull(attorney, v) {
			match = append(match, actor.TypeReplacementAttorney)
		}
	}

	if names.EqualFull(donor.CertificateProvider, v) {
		match = append(match, actor.TypeCertificateProvider)
	}

	for _, person := range donor.PeopleToNotify {
		if names.EqualFull(person, v) {
			match = append(match, actor.TypePersonToNotify)
		}
	}

	if names.EqualFull(donor.AuthorisedSignatory, v) {
		match = append(match, actor.TypeAuthorisedSignatory)
	}

	if names.EqualFull(donor.IndependentWitness, v) {
		match = append(match, actor.TypeIndependentWitness)
	}

	return match
}
