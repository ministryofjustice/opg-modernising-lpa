package actor

//go:generate enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypePersonalWelfare    LpaType = iota + 1 // personal-welfare
	LpaTypePropertyAndAffairs                    // property-and-affairs
)

func (e LpaType) LegacyString() string {
	switch e {
	case LpaTypePropertyAndAffairs:
		return "pfa"
	case LpaTypePersonalWelfare:
		return "hw"
	}
	return ""
}

func (e LpaType) LegalTermTransKey() string {
	switch e {
	case LpaTypePropertyAndAffairs:
		return "pfaLegalTerm"
	case LpaTypePersonalWelfare:
		return "hwLegalTerm"
	}
	return ""
}

func (e LpaType) WhatLPACoversTransKey() string {
	switch e {
	case LpaTypePropertyAndAffairs:
		return "whatPropertyAndAffairsCovers"
	case LpaTypePersonalWelfare:
		return "whatPersonalWelfareCovers"
	}
	return ""
}
