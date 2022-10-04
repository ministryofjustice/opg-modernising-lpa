package page

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadDate(t *testing.T) {
	date := readDate(time.Date(2020, time.March, 12, 0, 0, 0, 0, time.Local))

	assert.Equal(t, Date{Day: "12", Month: "3", Year: "2020"}, date)
}

func TestReadIdentityOption(t *testing.T) {
	assert.Equal(t, Passport, readIdentityOption("passport"))
	assert.Equal(t, IdentityOptionUnknown, readIdentityOption("what"))
}

func TestIdentityOptionsRanked(t *testing.T) {
	first, second := identityOptionsRanked([]IdentityOption{Passport, DrivingLicence, DwpAccount})
	assert.Equal(t, first, Passport)
	assert.Equal(t, second, DrivingLicence)

	first, second = identityOptionsRanked([]IdentityOption{GovernmentGatewayAccount, CouncilTaxBill, UtilityBill})
	assert.Equal(t, first, GovernmentGatewayAccount)
	assert.Equal(t, second, UtilityBill)

	first, second = identityOptionsRanked([]IdentityOption{DrivingLicence, OnlineBankAccount, GovernmentGatewayAccount})
	assert.Equal(t, first, DrivingLicence)
	assert.Equal(t, second, GovernmentGatewayAccount)
}
