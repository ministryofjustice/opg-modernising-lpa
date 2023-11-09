package donor

import (
	"testing"

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
	assert.Equal(t, &chooseAddressData{
		Form:      &form.AddressForm{},
		TitleKeys: testTitleKeys,
		App:       testAppData,
	}, newChooseAddressData(testAppData))
}

func TestOverrideProfessionalCertificateProviderKeys(t *testing.T) {
	data := newChooseAddressData(testAppData)
	data.overrideProfessionalCertificateProviderKeys()

	assert.Equal(t, &chooseAddressData{
		App:  testAppData,
		Form: &form.AddressForm{},
		TitleKeys: titleKeys{
			Manual:                          "personsWorkAddress",
			Postcode:                        "whatIsPersonsWorkPostcode",
			PostcodeSelectAndPostcodeLookup: "selectPersonsWorkAddress",
			ReuseAndReuseSelect:             "selectAnAddressForPerson",
			ReuseOrNew:                      "addPersonsAddress",
		},
	}, data)
}
