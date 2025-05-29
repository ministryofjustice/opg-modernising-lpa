package actor

import (
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type SameNameWarning struct {
	actor      Type
	matches    Type
	firstNames string
	lastName   string
	fullName   string
}

func NewSameNameWarning(actor, matches Type, fullName string) *SameNameWarning {
	if actor == TypeNone || matches == TypeNone {
		return nil
	}

	return &SameNameWarning{
		actor:    actor,
		matches:  matches,
		fullName: fullName,
	}
}

func (w *SameNameWarning) Format(l localize.Localizer) string {
	return l.Format(w.translationKey(), map[string]any{
		"Type":                l.T(w.actor.String()),
		"TypePlural":          strings.ToLower(l.T(w.pluralActorType())),
		"ArticleAndType":      l.T(w.actorArticleType()),
		"Match":               l.T(w.matches.String()),
		"MatchArticleAndType": l.T(w.matchesArticleType()),
		"FullName":            w.fullName,
	})
}

func (w *SameNameWarning) translationKey() string {
	switch w.actor {
	case TypeDonor:
		if w.matches.IsAttorney() || w.matches.IsReplacementAttorney() || w.matches.IsIndependentWitness() || w.matches.IsAuthorisedSignatory() {
			return "donorMatchesActorNameWarning"
		}

		if w.matches.IsCertificateProvider() {
			return "donorMatchesActorNameOrAddressWarning"
		}

		if w.matches.IsCorrespondent() {
			return "correspondentMatchesDonorNameWarning"
		}

		if w.matches.IsPersonToNotify() {
			return "personToNotifyMatchesDonorNameWarning"
		}
	case TypeCertificateProvider:
		if w.matches.IsDonor() {
			return "actorMatchesDonorNameOrAddressWarning"
		}

		return "actorMatchesDifferentActorNameOrAddressWarningConfirmLater"
	case TypeCorrespondent:
		if w.matches.IsDonor() {
			return "correspondentMatchesDonorNameWarning"
		}
	case TypePersonToNotify:
		if w.matches.IsPersonToNotify() {
			return "actorMatchesSameActorTypeNameWarning"
		}

		if w.matches.IsDonor() {
			return "personToNotifyMatchesDonorNameWarning"
		}

		if w.matches.IsAttorney() {
			return "personToNotifyMatchesAttorneyNameWarning"
		}

		// TODO Check with Abbi/Laura, do we want to show actorMatchesDifferentActorTypeNameWarning here?
		//return "actorMatchesDifferentActorTypeNameWarning"
	case TypeAttorney, TypeReplacementAttorney, TypeIndependentWitness, TypeAuthorisedSignatory:
		if w.matches == w.actor {
			return "actorMatchesSameActorTypeNameWarning"
		}

		if w.matches.IsDonor() {
			return "actorMatchesDonorNameWarning"
		}

		return "actorMatchesDifferentActorTypeNameWarning"
	}

	return ""
}

func (w *SameNameWarning) actorArticleType() string {
	return articleAndType(w.actor)
}

func (w *SameNameWarning) matchesArticleType() string {
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
	case TypeCorrespondent:
		return "theCorrespondent"
	}

	return ""
}

func (w *SameNameWarning) pluralActorType() string {
	switch w.actor {
	case TypeAttorney:
		return "attorneys"
	case TypeReplacementAttorney:
		return "replacementAttorneys"
	case TypeCertificateProvider:
		return "certificateProviders"
	case TypePersonToNotify:
		return "peopleToNotify"
	default:
		return ""
	}
}
