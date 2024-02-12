package donor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/stretchr/testify/assert"
)

var testTitleKeys = titleKeys{
	Manual:                          "personsAddress",
	Postcode:                        "whatIsPersonsPostcode",
	PostcodeSelectAndPostcodeLookup: "selectAnAddressForPerson",
	ReuseAndReuseSelect:             "selectAnAddressForPerson",
	ReuseOrNew:                      "addPersonsAddress",
}

func TestNewChooseAddressData(t *testing.T) {
	uid := actor.NewUID()

	assert.Equal(t, &chooseAddressData{
		Form:       &form.AddressForm{FieldNames: form.FieldNames.Address},
		TitleKeys:  testTitleKeys,
		App:        testAppData,
		ActorLabel: "a",
		FullName:   "b",
		UID:        uid,
		CanSkip:    true,
	}, newChooseAddressData(testAppData, "a", "b", uid, true))
}

func TestOverrideProfessionalCertificateProviderKeys(t *testing.T) {
	uid := actor.NewUID()
	data := newChooseAddressData(testAppData, "1", "2", uid, true)

	data.overrideTitleKeys(titleKeys{
		Manual:                          "a",
		PostcodeSelectAndPostcodeLookup: "b",
		Postcode:                        "c",
		ReuseAndReuseSelect:             "d",
		ReuseOrNew:                      "e",
	})

	assert.Equal(t, &chooseAddressData{
		Form: &form.AddressForm{FieldNames: form.FieldNames.Address},
		TitleKeys: titleKeys{
			Manual:                          "a",
			PostcodeSelectAndPostcodeLookup: "b",
			Postcode:                        "c",
			ReuseAndReuseSelect:             "d",
			ReuseOrNew:                      "e",
		},
		App:        testAppData,
		ActorLabel: "1",
		FullName:   "2",
		UID:        uid,
		CanSkip:    true,
	}, data)
}
