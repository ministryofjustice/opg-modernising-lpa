#!/usr/bin/env sh

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
             attorney-started \
             identity-check-mismatched \
             correspondent-updated \
             lpa-access-granted \
             letter-requested \
             confirm-at-post-office-selected \
             register-with-court-of-protection
do
    echo $v
    curl -o internal/event/testdata/$v.json "https://raw.githubusercontent.com/ministryofjustice/opg-event-store/main/domains/POAS/events/$v/schema.json"
done
