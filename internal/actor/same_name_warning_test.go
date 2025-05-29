package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslationKey(t *testing.T) {
	testcases := map[string]struct {
		warning             SameNameWarning
		expectedTranslation string
	}{
		"donor matches attorney": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeAttorney,
			},
			expectedTranslation: "donorMatchesActorNameWarning",
		},
		"donor matches replacement attorney": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeReplacementAttorney,
			},
			expectedTranslation: "donorMatchesActorNameWarning",
		},
		"donor matches independent witness": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeIndependentWitness,
			},
			expectedTranslation: "donorMatchesActorNameWarning",
		},
		"donor matches authorised signatory": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeAuthorisedSignatory,
			},
			expectedTranslation: "donorMatchesActorNameWarning",
		},
		"donor matches certificate provider": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeCertificateProvider,
			},
			expectedTranslation: "donorMatchesActorNameOrAddressWarning",
		},
		"donor matches correspondent": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypeCorrespondent,
			},
			expectedTranslation: "correspondentMatchesDonorNameWarning",
		},
		"donor matches person to notify": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: TypePersonToNotify,
			},
			expectedTranslation: "personToNotifyMatchesDonorNameWarning",
		},
		"certificate provider matches donor": {
			warning: SameNameWarning{
				actor:   TypeCertificateProvider,
				matches: TypeDonor,
			},
			expectedTranslation: "actorMatchesDonorNameOrAddressWarning",
		},
		"certificate provider matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeCertificateProvider,
				matches: TypeAttorney,
			},
			expectedTranslation: "actorMatchesDifferentActorNameOrAddressWarningConfirmLater",
		},
		"correspondent matches donor": {
			warning: SameNameWarning{
				actor:   TypeCorrespondent,
				matches: TypeDonor,
			},
			expectedTranslation: "correspondentMatchesDonorNameWarning",
		},
		"correspondent matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeCorrespondent,
				matches: TypeAttorney,
			},
			expectedTranslation: "",
		},
		"person to notify matches person to notify": {
			warning: SameNameWarning{
				actor:   TypePersonToNotify,
				matches: TypePersonToNotify,
			},
			expectedTranslation: "actorMatchesSameActorTypeNameWarning",
		},
		"person to notify matches donor": {
			warning: SameNameWarning{
				actor:   TypePersonToNotify,
				matches: TypeDonor,
			},
			expectedTranslation: "personToNotifyMatchesDonorNameWarning",
		},
		"person to notify matches attorney": {
			warning: SameNameWarning{
				actor:   TypePersonToNotify,
				matches: TypeAttorney,
			},
			expectedTranslation: "personToNotifyMatchesAttorneyNameWarning",
		},
		"person to notify matches any other actor": {
			warning: SameNameWarning{
				actor:   TypePersonToNotify,
				matches: TypeReplacementAttorney,
			},
			expectedTranslation: "",
		},
		"attorney matches attorney": {
			warning: SameNameWarning{
				actor:   TypeAttorney,
				matches: TypeAttorney,
			},
			expectedTranslation: "actorMatchesSameActorTypeNameWarning",
		},
		"attorney matches donor": {
			warning: SameNameWarning{
				actor:   TypeAttorney,
				matches: TypeDonor,
			},
			expectedTranslation: "actorMatchesDonorNameWarning",
		},
		"attorney matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeAttorney,
				matches: TypeCertificateProvider,
			},
			expectedTranslation: "actorMatchesDifferentActorTypeNameWarning",
		},
		"replacement attorney matches replacement attorney": {
			warning: SameNameWarning{
				actor:   TypeReplacementAttorney,
				matches: TypeReplacementAttorney,
			},
			expectedTranslation: "actorMatchesSameActorTypeNameWarning",
		},
		"replacement attorney matches donor": {
			warning: SameNameWarning{
				actor:   TypeReplacementAttorney,
				matches: TypeDonor,
			},
			expectedTranslation: "actorMatchesDonorNameWarning",
		},
		"replacement attorney matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeReplacementAttorney,
				matches: TypeCertificateProvider,
			},
			expectedTranslation: "actorMatchesDifferentActorTypeNameWarning",
		},
		"independent witness matches donor": {
			warning: SameNameWarning{
				actor:   TypeIndependentWitness,
				matches: TypeDonor,
			},
			expectedTranslation: "actorMatchesDonorNameWarning",
		},
		"independent witness matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeIndependentWitness,
				matches: TypeCertificateProvider,
			},
			expectedTranslation: "actorMatchesDifferentActorTypeNameWarning",
		},
		"authorised signatory matches donor": {
			warning: SameNameWarning{
				actor:   TypeAuthorisedSignatory,
				matches: TypeDonor,
			},
			expectedTranslation: "actorMatchesDonorNameWarning",
		},
		"authorised signatory matches any other actor": {
			warning: SameNameWarning{
				actor:   TypeAuthorisedSignatory,
				matches: TypeCertificateProvider,
			},
			expectedTranslation: "actorMatchesDifferentActorTypeNameWarning",
		},
		"unexpected type": {
			warning: SameNameWarning{
				actor:   TypeDonor,
				matches: Type(99),
			},
			expectedTranslation: "",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedTranslation, tc.warning.translationKey())
		})
	}

}
