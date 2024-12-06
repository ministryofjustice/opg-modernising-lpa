console.log(`Request - ${context.request.method} ${context.request.path}`);

const lpaStore = stores.open('lpa');
const pathParts = context.request.path.split('/');
const lpaUID = pathParts[2]

switch (context.request.method) {
    case 'GET': {
        if (pathParts.length == 3 && pathParts[1] == 'lpas') {
            const lpa = lpaStore.load(lpaUID);
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
        lpa.uid = lpaUID;
        lpa.updatedAt = new Date(Date.now()).toISOString();
        lpaStore.save(lpaUID, JSON.stringify(lpa));
        respond();
        break;
    }
    case 'POST': {
        if (context.request.path == '/lpas') {
            let uids = JSON.parse(context.request.body).uids;
            let lpas = uids.map(uid => lpaStore.load(uid)).reduce((list, lpa) => lpa ? list.concat([JSON.parse(lpa)]) : list, []);

            respond().withContent(JSON.stringify({ lpas: lpas }));
        } else {
            let update = JSON.parse(context.request.body);
            let lpa = JSON.parse(lpaStore.load(lpaUID));
            if (!lpa) {
                respond().withStatusCode(404);
                break;
            }
            lpa.updatedAt = new Date(Date.now()).toISOString();

            switch (update.type) {
                case 'ATTORNEY_SIGN': {
                    const keyParts = update.changes[0].key.split('/');
                    const idx = parseInt(keyParts[2]);
                    const signedAt = update.changes.find(x => x.key.includes('signedAt')).new;

                    if (lpa.attorneys && idx < lpa.attorneys.length) {
                        lpa.attorneys[idx].signedAt = signedAt;
                    }
                    break;
                }
                case 'TRUST_CORPORATION_SIGN': {
                    const keyParts = update.changes[0].key.split('/');
                    const idx = parseInt(keyParts[2]);
                    const signedAt = update.changes.find(x => x.key.includes('signedAt')).new;

                    if (lpa.trustCorporations && idx < lpa.trustCorporations.length) {
                        lpa.trustCorporations[idx].signatories = [{ firstNames: "A", lastName: "Sign", signedAt: signedAt }];
                    }
                    break;
                }
                case 'CERTIFICATE_PROVIDER_SIGN':
                    const signedAt = update.changes.find(x => x.key.includes('signedAt')).new;
                    lpa.certificateProvider.signedAt = signedAt;
                    break;

                case 'STATUTORY_WAITING_PERIOD':
                    lpa.status = 'statutory-waiting-period';
                    break;

                case 'REGISTER':
                    if (lpa.status === 'statutory-waiting-period') {
                        lpa.status = 'registered';
                        lpa.registrationDate = new Date(Date.now()).toISOString();
                    }
                    break;

                case 'CERTIFICATE_PROVIDER_OPT_OUT':
                    lpa.status = 'cannot-register';
                    break;

                case 'DONOR_WITHDRAW_LPA':
                    lpa.status = 'withdrawn';
                    break;

                case 'ATTORNEY_OPT_OUT':
                    const idx = lpa.attorneys.findIndex(item => item.uid === update.subject)

                    if (idx >= 0 && lpa.attorneys[idx].signedAt != '') {
                        lpa.attorneys[idx].status = 'removed'
                    }
                    break;

                case 'OPG_STATUS_CHANGE':
                    lpa.status = update.changes.find(x => x.key == '/status').new;
                    break;
            }

            lpaStore.save(lpaUID, JSON.stringify(lpa));
            respond();
        }
        break;
    }
    default:
        respond();
}
