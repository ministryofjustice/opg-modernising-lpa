package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoucherFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Voucher{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestVoucherMatches(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{FirstNames: "a", LastName: "b"},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Nil(t, Voucher{FirstNames: "x", LastName: "y"}.Matches(donor))
	assert.Equal(t, []Type{TypeDonor}, Voucher{FirstNames: "a", LastName: "b"}.Matches(donor))
	assert.Equal(t, []Type{TypeAttorney}, Voucher{FirstNames: "C", LastName: "D"}.Matches(donor))
	assert.Equal(t, []Type{TypeAttorney}, Voucher{FirstNames: "e", LastName: "f"}.Matches(donor))
	assert.Equal(t, []Type{TypeReplacementAttorney}, Voucher{FirstNames: "G", LastName: "H"}.Matches(donor))
	assert.Equal(t, []Type{TypeReplacementAttorney}, Voucher{FirstNames: "i", LastName: "j"}.Matches(donor))
	assert.Equal(t, []Type{TypeCertificateProvider}, Voucher{FirstNames: "k", LastName: "l"}.Matches(donor))
	assert.Equal(t, []Type{TypePersonToNotify}, Voucher{FirstNames: "m", LastName: "n"}.Matches(donor))
	assert.Equal(t, []Type{TypePersonToNotify}, Voucher{FirstNames: "O", LastName: "P"}.Matches(donor))
	assert.Equal(t, []Type{TypeAuthorisedSignatory}, Voucher{FirstNames: "a", LastName: "s"}.Matches(donor))
	assert.Equal(t, []Type{TypeIndependentWitness}, Voucher{FirstNames: "i", LastName: "w"}.Matches(donor))
}

func TestVoucherMatchesMultiple(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{FirstNames: "a", LastName: "b"},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "a", LastName: "b"},
			{FirstNames: "a", LastName: "b"},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "a", LastName: "b"},
			{FirstNames: "a", LastName: "b"},
		}},
		CertificateProvider: CertificateProvider{FirstNames: "a", LastName: "b"},
		PeopleToNotify: PeopleToNotify{
			{FirstNames: "a", LastName: "b"},
			{FirstNames: "a", LastName: "b"},
		},
		AuthorisedSignatory: AuthorisedSignatory{FirstNames: "a", LastName: "b"},
		IndependentWitness:  IndependentWitness{FirstNames: "a", LastName: "b"},
	}

	assert.Equal(t, []Type{TypeDonor, TypeAttorney, TypeAttorney, TypeReplacementAttorney, TypeReplacementAttorney,
		TypeCertificateProvider, TypePersonToNotify, TypePersonToNotify, TypeAuthorisedSignatory, TypeIndependentWitness},
		Voucher{FirstNames: "a", LastName: "b"}.Matches(donor))
}

func TestVoucherMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{FirstNames: "", LastName: ""},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Nil(t, Voucher{FirstNames: "", LastName: ""}.Matches(donor))
}
