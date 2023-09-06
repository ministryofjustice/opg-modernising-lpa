
var paymentsStore = stores.open('payments');

switch (context.request.method) {
    case 'GET':
        console.log(paymentsStore.load('amount'))
        var response = respond()
        var respBody = JSON.parse(response.content)

        respBody.amount = paymentsStore.load('amount')
        respond().withContent(JSON.stringify(respBody))

        break
    case 'POST':
        var reqBody = JSON.parse(context.request.body)

        paymentsStore.save('amount', reqBody.amount)
        respond()

        break
    default:
        respond()
}
