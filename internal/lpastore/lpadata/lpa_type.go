package lpadata

//go:generate enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypePersonalWelfare    LpaType = iota + 1 // personal-welfare
	LpaTypePropertyAndAffairs                    // property-and-affairs
)

// TODO: move this somewhere else
func (e LpaType) WhatLPACoversTransKey() string {
	switch e {
	case LpaTypePropertyAndAffairs:
		return "whatPropertyAndAffairsCovers"
	case LpaTypePersonalWelfare:
		return "whatPersonalWelfareCovers"
	}
	return ""
}
