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
    lpaStore.save(pathParts[2], JSON.stringify(lpa));
    respond();
    break;
  }
  case 'POST': {
    let update = JSON.parse(context.request.body);
    let lpa = JSON.parse(lpaStore.load(pathParts[2]));

    console.log(JSON.stringify(update));
    console.log(JSON.stringify(lpa));
    switch (update.type) {
      case 'ATTORNEY_SIGN':
        const keyParts = update.changes[0].key.split('/');
        const idx = parseInt(keyParts[2]);

        switch (keyParts[1]) {
          case 'attorneys':
            if (lpa.attorneys && idx < lpa.attorneys.length) {
              lpa.attorneys[idx].signedAt = lpa.signedAt;
            }

          case 'trustCorporations':
            if (lpa.trustCorporations && idx < lpa.trustCorporations.length) {
              lpa.trustCorporations[idx].signatories = [{ signedAt: lpa.signedAt }];
            }
        }
    }

    lpaStore.save(pathParts[2], JSON.stringify(lpa));
    respond();
    break;
  }
  default:
    respond();
}
