package actor

import "strings"

type Voucher struct {
	FirstNames string
	LastName   string
	Email      string
	Allowed    bool
}

func (v Voucher) FullName() string {
	return v.FirstNames + " " + v.LastName
}

func (v Voucher) Matches(donor *DonorProvidedDetails) (match []Type) {
	if v.FirstNames == "" && v.LastName == "" {
		return nil
	}

	if strings.EqualFold(donor.Donor.FirstNames, v.FirstNames) && strings.EqualFold(donor.Donor.LastName, v.LastName) {
		match = append(match, TypeDonor)
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, v.FirstNames) && strings.EqualFold(attorney.LastName, v.LastName) {
			match = append(match, TypeAttorney)
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, v.FirstNames) && strings.EqualFold(attorney.LastName, v.LastName) {
			match = append(match, TypeReplacementAttorney)
		}
	}

	if strings.EqualFold(donor.CertificateProvider.FirstNames, v.FirstNames) && strings.EqualFold(donor.CertificateProvider.LastName, v.LastName) {
		match = append(match, TypeCertificateProvider)
	}

	for _, person := range donor.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, v.FirstNames) && strings.EqualFold(person.LastName, v.LastName) {
			match = append(match, TypePersonToNotify)
		}
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, v.FirstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, v.LastName) {
		match = append(match, TypeAuthorisedSignatory)
	}

	if strings.EqualFold(donor.IndependentWitness.FirstNames, v.FirstNames) && strings.EqualFold(donor.IndependentWitness.LastName, v.LastName) {
		match = append(match, TypeIndependentWitness)
	}

	return match
}
