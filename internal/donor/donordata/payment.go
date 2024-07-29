package donordata

type Payment struct {
	// Reference generated for the payment
	PaymentReference string
	// ID returned from GOV.UK Pay
	PaymentId string
	// Amount is the amount paid in pence
	Amount int
}
