# Changes to existing GOV.UK Notify SMS and email templates

Email and SMS notifications are sent via [GOV.UK Notify](https://www.notifications.service.gov.uk). If a change to an existing template involves removing or adding a personalisation variable we should take a staged approach to making the change to avoid missing personalisation error responses from Notify in production or ephemeral environments.

- Copy the existing template in Notify:
  - Templates `->` New Template `->` Copy an existing template
- Make required changes, move the template to a suitable folder and copy the new template ID
- Update the [template ID](../../internal/notify/client.go) in code as part of the PR
- Merge the PR
- Remove the old template from Notify

If there is an existing template in the OPG Test Notify account that needs to be duplicated you can copy this directly in to the Office of the Public Guardian account by following the steps above but selecting the template from OPG Test when choosing Copy an existing template.

This process should be followed in both OPG Test and Office of the Public Guardian Notify accounts.
