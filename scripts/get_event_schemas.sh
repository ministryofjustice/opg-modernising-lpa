#!/usr/bin/env sh

mkdir -p internal/event/testdata
curl -o internal/event/testdata/application-updated.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/application-updated/schema.json"
curl -o internal/event/testdata/reduced-fee-requested.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/reduced-fee-requested/schema.json"
curl -o internal/event/testdata/notification-sent.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/notification-sent/schema.json"
curl -o internal/event/testdata/paper-form-requested.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/paper-form-requested/schema.json"
