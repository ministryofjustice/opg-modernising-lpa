def lpaStore = stores.open('lpa')

switch (context.request.method) {
    case 'GET':
        def parts = context.request.path.split('/')
        logger.warn("UID: " + parts[2])
        def lpa = lpaStore.load(parts[2])
        if (lpa) {
            logger.warn("FOUND")
            respond().withContent(lpa)
        } else {
            logger.warn("NOT FOUND")
            respond()
        }
        break

    case 'PUT':
        def parts = context.request.path.split('/')
        logger.warn("UID: " + parts[2])
        lpaStore.save(parts[2], context.request.body)
        respond()
        break

    default:
        respond()
}
