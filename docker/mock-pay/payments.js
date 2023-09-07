var paymentsStore = stores.open('payments');

switch (context.request.method) {
    case 'GET':
        console.log(paymentsStore.load('amount'))
        if (paymentsStore.load('amount') === 4100) {
            respond().withExampleName('half-fee')
        } else {
            respond().withExampleName('full-fee')
        }

        break
    case 'POST':
        var reqBody = JSON.parse(context.request.body)

        paymentsStore.save('amount', reqBody.amount)
        respond()

        break
    default:
        respond()
}
