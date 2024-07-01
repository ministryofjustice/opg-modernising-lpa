# Changes to existing GOV.UK Notify SMS and email templates

Email and SMS notifications are sent via [GOV.UK Notify](https://www.notifications.service.gov.uk). If a change to an existing template involves removing or adding a personalisation variable we should take a staged approach to making the change to avoid missing personalisation error responses from Notify in production or ephemeral environments.

- Ensure you are viewing the `Office of the Public Guardian` (visible in top left). If not, click `Switch service` in top right and select `Office of the Public Guardian`
- Click `Templates` and find the existing template in `Modernising LPA -> Live`
- Copy the existing template in Notify:
  - `New Template -> Copy an existing template`
- Make required changes, move the template to `Modernising LPA -> Staging` and copy the new template ID
- Update the [template ID](../../internal/notify/client.go) in code as part of the PR
- Merge the PR
- Remove the old template from Notify
- Move the template from `Modernising LPA -> Staging` to `Modernising LPA -> Live`

This process should be followed in both `OPG Test` and `Office of the Public Guardian` Notify services.

If there is an existing template in the `OPG Test` Notify service that needs to be duplicated you can make a copy of it directly by switching to the `Office of the Public Guardian` Notify service and then following the steps above but selecting the template from `OPG Test` when choosing `Copy an existing template`.


