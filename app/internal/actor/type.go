package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type Type int

const (
	TypeNone Type = iota
	TypeDonor
	TypeAttorney
	TypeReplacementAttorney
	TypeCertificateProvider
	TypePersonToNotify
)

type SameNameWarning struct {
	actor      Type
	matches    Type
	firstNames string
	lastName   string
}

func NewSameNameWarning(actor, matches Type, firstNames, lastName string) *SameNameWarning {
	if actor == TypeNone || matches == TypeNone {
		return nil
	}

	return &SameNameWarning{
		actor:      actor,
		matches:    matches,
		firstNames: firstNames,
		lastName:   lastName,
	}
}

func (w *SameNameWarning) Key() string {
	switch w.matches {
	case TypeDonor:
		return "donorMatchesActorWarning"
	case TypeAttorney:
		if w.actor == TypeAttorney {
			return "attorneyMatchesAttorneyWarning"
		}
		return "attorneyMatchesActorWarning"
	case TypeReplacementAttorney:
		if w.actor == TypeReplacementAttorney {
			return "replacementAttorneyMatchesReplacementAttorneyWarning"
		}
		return "replacementAttorneyMatchesActorWarning"
	case TypeCertificateProvider:
		return "certificateProviderMatchesActorWarning"
	case TypePersonToNotify:
		if w.actor == TypePersonToNotify {
			return "personToNotifyMatchesPersonToNotifyWarning"
		}
		return "personToNotifyMatchesActorWarning"
	}

	return ""
}

func (w *SameNameWarning) Type() string {
	switch w.actor {
	case TypeDonor:
		return "theDonor"
	case TypeAttorney:
		return "anAttorney"
	case TypeReplacementAttorney:
		return "aReplacementAttorney"
	case TypeCertificateProvider:
		return "theCertificateProvider"
	case TypePersonToNotify:
		return "aPersonToNotify"
	}

	return ""
}

func (w *SameNameWarning) Format(l localize.Localizer) string {
	return l.Format(w.Key(), map[string]any{
		"Type":       l.T(w.Type()),
		"FirstNames": w.firstNames,
		"LastName":   w.lastName,
	})
}

func (w *SameNameWarning) String() string {
	if w == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%d|%d|%s|%s", w.actor, w.matches, w.firstNames, w.lastName)
}
