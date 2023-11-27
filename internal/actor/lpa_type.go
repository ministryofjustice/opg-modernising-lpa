package actor

//go:generate enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypeHealthWelfare   LpaType = iota + 1 // hw
	LpaTypePropertyFinance                    // pfa
)

func (e LpaType) LegalTermTransKey() string {
	switch e {
	case LpaTypePropertyFinance:
		return "pfaLegalTerm"
	case LpaTypeHealthWelfare:
		return "hwLegalTerm"
	}
	return ""
}

func (e LpaType) WhatLPACoversTransKey() string {
	switch e {
	case LpaTypePropertyFinance:
		return "whatPersonalAffairsCovers"
	case LpaTypeHealthWelfare:
		return "whatPersonalWelfareCovers"
	}
	return ""
}
