def lpaStore = stores.open('lpa')

switch (context.request.method) {
    case 'GET':
        def parts = context.request.path.split('/')
        def lpa = lpaStore.load(parts[2])
        if (lpa) {
            respond().withContent(lpa)
        } else {
            respond()
        }
        break

    case 'PUT':
        def parts = context.request.path.split('/')
        lpaStore.save(parts[2], context.request.body)
        respond()
        break

    default:
        respond()
}
