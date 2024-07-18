const paymentsStore = stores.open('payments');

switch (context.request.method) {
    case 'GET':
        const getPaymentResponseBody = `{
    "amount": 8200,
    "description": "Property and Finance LPA",
    "reference": "Hxzqvk78fBdl",
    "language": "en",
    "email": "simulate-delivered@notifications.service.gov.uk",
    "state": {
        "status": "success",
        "finished": true
    },
    "payment_id": "7o5rc438t2f1sv4fs3pome24ju",
    "payment_provider": "sandbox",
    "created_date": "2024-07-17T14:35:35.085Z",
    "refund_summary": {
        "status": "available",
        "amount_available": 8200,
        "amount_submitted": 0
    },
    "settlement_summary": {
        "capture_submit_time": "2024-07-17T14:36:25.896Z",
        "captured_date": "2024-07-17"
    },
    "card_details": {
        "last_digits_card_number": "1111",
        "first_digits_card_number": "444433",
        "cardholder_name": "Mr Sam Smith",
        "expiry_date": "01/27",
        "billing_address": {
            "line1": "1 RICHMOND PLACE",
            "line2": "KINGS HEATH",
            "postcode": "B14 7ED",
            "city": "BIRMINGHAM",
            "country": "GB"
        },
        "card_brand": "Visa",
        "card_type": "credit",
        "wallet_type": null
    },
    "delayed_capture": false,
    "moto": false,
    "provider_id": "7fb768f5-939c-4264-b7f3-e0e482e7c175",
    "return_url": "http://localhost:5050/lpa/82921d1f-e6fa-40a3-9f3d-0879bf334e13/payment-confirmation",
    "authorisation_mode": "web",
    "_links": {
        "self": {
            "href": "https://publicapi.payments.service.gov.uk/v1/payments/7o5rc438t2f1sv4fs3pome24ju",
            "method": "GET"
        },
        "events": {
            "href": "https://publicapi.payments.service.gov.uk/v1/payments/7o5rc438t2f1sv4fs3pome24ju/events",
            "method": "GET"
        },
        "refunds": {
            "href": "https://publicapi.payments.service.gov.uk/v1/payments/7o5rc438t2f1sv4fs3pome24ju/refunds",
            "method": "GET"
        }
    },
    "card_brand": "Visa"
}
`

        const payment = JSON.parse(paymentsStore.load('payment'))
        let response = JSON.parse(getPaymentResponseBody)
        let now = new Date()

        response.amount = payment.amount
        response.email = payment.email
        response.description = payment.description
        response.reference = payment.reference
        response.refund_summary.amount_available = payment.amount
        response.settlement_summary.capture_submit_time = now.toISOString()
        response.settlement_summary.captured_date = now.toISOString().split('T')[0]

        now.setMinutes(now.getMinutes()-1)
        response.created_date = now.toISOString()

        respond().withContent(JSON.stringify(response))

        break
    case 'POST':
        paymentsStore.save('payment', context.request.body)

        respond()

        break
    default:
        respond()
}
