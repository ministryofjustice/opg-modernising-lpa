package pay

import (
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

type CreatePaymentBody struct {
	Amount      int    `json:"amount"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
	ReturnURL   string `json:"return_url"`
	Email       string `json:"email"`
	Language    string `json:"language"`
}

type State struct {
	Status   string `json:"status"`
	Finished bool   `json:"finished"`
}

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type CreatePaymentResponse struct {
	CreatedDate     time.Time       `json:"created_date"`
	State           State           `json:"State"`
	Links           map[string]Link `json:"_links"`
	Amount          int             `json:"amount"`
	Reference       string          `json:"reference"`
	Description     string          `json:"description"`
	ReturnURL       string          `json:"return_url"`
	PaymentID       string          `json:"payment_id"`
	PaymentProvider string          `json:"payment_provider"`
	ProviderID      string          `json:"provider_id"`
}

type CardDetails struct {
	CardBrand             string         `json:"card_brand"`
	CardType              string         `json:"card_type"`
	LastDigitsCardNumber  string         `json:"last_digits_card_number"`
	FirstDigitsCardNumber string         `json:"first_digits_card_number"`
	ExpiryDate            string         `json:"expiry_date"`
	CardholderName        string         `json:"cardholder_name"`
	BillingAddress        BillingAddress `json:"billing_address"`
}

type BillingAddress struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2"`
	Postcode string `json:"postcode"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

type AuthorisationSummary struct {
	ThreeDSecure ThreeDSecure `json:"three_d_secure"`
}

type ThreeDSecure struct {
	Required bool `json:"required"`
}

type RefundSummary struct {
	Status          string      `json:"status"`
	AmountAvailable AmountPence `json:"amount_available"`
}

type SettlementSummary struct {
	CaptureSubmitTime time.Time `json:"capture_submit_time"`
	CapturedDate      date.Date `json:"captured_date"`
	SettledDate       date.Date `json:"settled_date"`
}

type AmountPence int

func (a AmountPence) String() string {
	return "£" + humanize.CommafWithDigits(float64(a)/100, 2)
}

func (a AmountPence) Pence() int {
	return int(a)
}

type GetPaymentResponse struct {
	CreatedDate time.Time   `json:"created_date"`
	AmountPence AmountPence `json:"amount"`
	State       State       `json:"State"`
	Description string      `json:"description"`
	Reference   string      `json:"reference"`
	Language    string      `json:"language"`
	//May be useful but until we define if/what we send in CreatePayment we can't marshal the response
	//
	//Metadata    struct {
	//	LedgerCode                string `json:"ledger_code"`
	//	AnInternalReferenceNumber int    `json:"an_internal_reference_number"`
	//} `json:"metadata"`
	Email                  string               `json:"email"`
	CardDetails            CardDetails          `json:"card_details"`
	PaymentID              string               `json:"payment_id"`
	AuthorisationSummary   AuthorisationSummary `json:"authorisation_summary"`
	RefundSummary          RefundSummary        `json:"refund_summary"`
	SettlementSummary      SettlementSummary    `json:"settlement_summary"`
	DelayedCapture         bool                 `json:"delayed_capture"`
	Moto                   bool                 `json:"moto"`
	CorporateCardSurcharge int                  `json:"corporate_card_surcharge"`
	TotalAmount            int                  `json:"total_amount"`
	Fee                    int                  `json:"fee"`
	NetAmount              int                  `json:"net_amount"`
	PaymentProvider        string               `json:"payment_provider"`
	ProviderID             string               `json:"provider_id"`
	ReturnURL              string               `json:"return_url"`
}
