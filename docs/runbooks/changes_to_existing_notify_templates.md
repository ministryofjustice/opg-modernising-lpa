# Changes to existing GOV.UK Notify SMS and email templates

Email and SMS notifications are sent via [GOV.UK Notify](https://www.notifications.service.gov.uk). If a change to an existing template involves removing or adding a personalisation variable we should take a staged approach to making the change to avoid missing personalisation error responses from Notify in production or ephemeral environments.

- Ensure you are viewing the `Office of the Public Guardian` (visible in top left). If not, click `Switch service` in top right and select `Office of the Public Guardian`
- Click `Templates` and find the existing template in `Modernising LPA -> Live`
- Copy the existing template in Notify:
  - `New Template -> Copy an existing template`
- Make required changes, move the template to `Modernising LPA -> Staging` and copy the new template ID
- Update the [email template ID](../../internal/notify/email.go) or [SMS template ID](../../internal/notify/sms.go) in code as part of the PR
- Merge the PR and wait for changes to apply on production
- Remove the old template from Notify
- Move the template from `Modernising LPA -> Staging` to `Modernising LPA -> Live`
