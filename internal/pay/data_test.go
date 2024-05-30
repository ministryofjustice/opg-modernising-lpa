package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPaymentURL(t *testing.T) {
	assert.True(t, IsPaymentURL("https://www.payments.service.gov.uk/whatever?hey"))
	assert.True(t, IsPaymentURL("https://card.payments.service.gov.uk/whatever?hey"))

	assert.False(t, IsPaymentURL("https://card.payments.service.gov.co/whatever?hey"))
	assert.False(t, IsPaymentURL("http://card.payments.service.gov.uk/whatever?hey"))
}
