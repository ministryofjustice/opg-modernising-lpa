package donordata

import (
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/temporary"
)

type Voucher struct {
	FirstNames string
	LastName   string
	Email      string
	Allowed    bool
}

func (v Voucher) FullName() string {
	return v.FirstNames + " " + v.LastName
}

func (v Voucher) Matches(donor *Provided) (match []temporary.ActorType) {
	if v.FirstNames == "" && v.LastName == "" {
		return nil
	}

	if strings.EqualFold(donor.Donor.FirstNames, v.FirstNames) && strings.EqualFold(donor.Donor.LastName, v.LastName) {
		match = append(match, temporary.ActorTypeDonor)
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, v.FirstNames) && strings.EqualFold(attorney.LastName, v.LastName) {
			match = append(match, temporary.ActorTypeAttorney)
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, v.FirstNames) && strings.EqualFold(attorney.LastName, v.LastName) {
			match = append(match, temporary.ActorTypeReplacementAttorney)
		}
	}

	if strings.EqualFold(donor.CertificateProvider.FirstNames, v.FirstNames) && strings.EqualFold(donor.CertificateProvider.LastName, v.LastName) {
		match = append(match, temporary.ActorTypeCertificateProvider)
	}

	for _, person := range donor.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, v.FirstNames) && strings.EqualFold(person.LastName, v.LastName) {
			match = append(match, temporary.ActorTypePersonToNotify)
		}
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, v.FirstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, v.LastName) {
		match = append(match, temporary.ActorTypeAuthorisedSignatory)
	}

	if strings.EqualFold(donor.IndependentWitness.FirstNames, v.FirstNames) && strings.EqualFold(donor.IndependentWitness.LastName, v.LastName) {
		match = append(match, temporary.ActorTypeIndependentWitness)
	}

	return match
}
