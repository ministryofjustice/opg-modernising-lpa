package actor

import (
	"fmt"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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

func (w *SameNameWarning) Format(l localize.Localizer) string {
	return l.Format(w.translationKey(), map[string]any{
		"Type":                l.T(w.actor.String()),
		"TypePlural":          strings.ToLower(l.T(w.pluralActorType())),
		"ArticleAndType":      l.T(w.actorType()),
		"MatchArticleAndType": l.T(w.matchesType()),
		"FirstNames":          w.firstNames,
		"LastName":            w.lastName,
	})
}

func (w *SameNameWarning) String() string {
	if w == nil {
		return "<nil>"
	}

	return fmt.Sprintf("%d|%d|%s|%s", w.actor, w.matches, w.firstNames, w.lastName)
}

func (w *SameNameWarning) translationKey() string {
	switch w.matches {
	case TypeDonor:
		return "donorMatchesActorWarning"
	case TypeAttorney:
		if w.actor == TypeAttorney {
			return "actorMatchesSameActorTypeWarning"
		}
		return "actorMatchesDifferentActorTypeWarning"
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
	case TypeAuthorisedSignatory:
		return "authorisedSignatoryMatchesActorWarning"
	case TypeIndependentWitness:
		return "independentWitnessMatchesActorWarning"
	}

	return ""
}

func (w *SameNameWarning) actorType() string {
	return articleAndType(w.actor)
}

func (w *SameNameWarning) matchesType() string {
	return articleAndType(w.matches)
}

func articleAndType(comparator Type) string {
	switch comparator {
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
	case TypeAuthorisedSignatory:
		return "theAuthorisedSignatory"
	case TypeIndependentWitness:
		return "theIndependentWitness"
	case TypeVoucher:
		return "theVoucher"
	}

	return ""
}

func (w *SameNameWarning) pluralActorType() string {
	switch w.actor {
	case TypeAttorney:
		return "attorneys"
	default:
		return ""
	}
}
