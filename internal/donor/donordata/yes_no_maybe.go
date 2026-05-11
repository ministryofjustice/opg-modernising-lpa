package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

//go:generate go tool enumerator -type YesNoMaybe -linecomment -empty
type YesNoMaybe uint8

const (
	Yes YesNoMaybe = iota + 1
	No
	Maybe
)

type YesNoMaybeForm = newforms.EnumForm[YesNoMaybe, YesNoMaybeOptions, *YesNoMaybe]

func NewYesNoMaybeForm(errorLabel string) *YesNoMaybeForm {
	return newforms.NewEnumForm[YesNoMaybe](errorLabel, YesNoMaybeValues)
}
