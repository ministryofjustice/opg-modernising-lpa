#!/usr/bin/env sh

mkdir -p internal/app/testdata
curl -o internal/app/testdata/application-updated.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/application-updated/schema.json"
curl -o internal/app/testdata/reduced-fee-requested.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/reduced-fee-requested/schema.json"
