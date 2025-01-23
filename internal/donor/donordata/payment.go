package donordata

import "time"

type Payment struct {
	// Reference generated for the payment
	PaymentReference string
	// ID returned from GOV.UK Pay
	PaymentID string
	// Amount is the amount paid in pence
	Amount int
	// CreatedAt is when the payment was created
	CreatedAt time.Time
}
