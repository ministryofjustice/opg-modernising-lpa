#!/usr/bin/env bash

mkdir -p internal/event/testdata
rm -f internal/event/testdata/*

for v in uid-requested \
             application-deleted \
             application-updated \
             reduced-fee-requested \
             notification-sent \
             paper-form-requested \
             payment-received \
             certificate-provider-started \
             attorney-started 
do
    echo $v
    curl -o internal/event/testdata/$v.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/$v/schema.json"
done
