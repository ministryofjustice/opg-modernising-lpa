console.log(`Request - ${context.request.method} ${context.request.path}`);

const lpaStore = stores.open('lpa');
const pathParts = context.request.path.split('/');

switch (context.request.method) {
case 'GET': {
    if (pathParts.length == 3 && pathParts[1] == 'lpas') {
        const lpa = lpaStore.load(pathParts[2]);
        if (lpa) {
            respond().withContent(lpa);
        } else {
            respond().withStatusCode(404);
        }
    } else {
        respond();
    }
    break;
}
case 'PUT': {
    let lpa = JSON.parse(context.request.body);
    lpa.uid = pathParts[2];
    lpa.updatedAt = new Date(Date.now()).toISOString();
    lpaStore.save(pathParts[2], JSON.stringify(lpa));
    respond();
    break;
}
case 'POST': {
    if (context.request.path == '/lpas') {
        let uids = JSON.parse(context.request.body).uids;
        let lpas = uids.map(uid => lpaStore.load(uid)).reduce((a,e) => e ? a.concat([JSON.parse(e)]) : a, []);

        respond().withContent(JSON.stringify({lpas: lpas}));
    } else {
        let update = JSON.parse(context.request.body);
        let lpa = JSON.parse(lpaStore.load(pathParts[2]));
        lpa.updatedAt = new Date(Date.now()).toISOString();

        switch (update.type) {
        case 'ATTORNEY_SIGN': {
            const keyParts = update.changes[0].key.split('/');
            const idx = parseInt(keyParts[2]);

            if (lpa.attorneys && idx < lpa.attorneys.length) {
                lpa.attorneys[idx].signedAt = lpa.signedAt;
            }
            break;
        }
        case 'TRUST_CORPORATION_SIGN': {
            const keyParts = update.changes[0].key.split('/');
            const idx = parseInt(keyParts[2]);

            if (lpa.trustCorporations && idx < lpa.trustCorporations.length) {
                lpa.trustCorporations[idx].signatories = [{ signedAt: lpa.signedAt }];
            }
            break;
        }
        case 'CERTIFICATE_PROVIDER_SIGN':
            lpa.certificateProvider.signedAt = lpa.signedAt;
            break;

        case 'REGISTER':
            lpa.status = 'registered';
            lpa.registrationDate = new Date(Date.now()).toISOString();
            break;
        }

        lpaStore.save(pathParts[2], JSON.stringify(lpa));
        respond();
    }
    break;
}
default:
    respond();
}
