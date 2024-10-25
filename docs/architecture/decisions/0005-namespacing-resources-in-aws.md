# 3. Namespacing resources in AWS

Date: 2024-06-19

## Status

Accepted

## Context

This decision is in relation to the name attribute of a resource, not the resource name itself. So for the example below, the name attribute is `event-received-${data.aws_default_tags.current.tags.environment-name}` and the resource name is `event-received`.

```hcl
resource "aws_iam_role" "event_received" {
  name = "event-received-${data.aws_default_tags.current.tags.environment-name}"
  ...
}
```

Making resources in AWS unique is important to avoid conflicts across environments, regions, and accounts. This is especially important when using resources that are shared across multiple environments such as encryption keys.

Granting access to resources is also easier when they are namespaced because it is easier to identify which resources are being accessed.

To make granting access to resources easier, we should use a consistent naming convention for resources.

The values currently used in naming resources are:

- `environment-name`
- `region`
- `resource-name`
- `account-name`
- `application-name`

Some examples of namespaced IAM role resources are:

- `event-received-${data.aws_default_tags.current.tags.environment-name}`
- `${data.aws_default_tags.current.tags.environment-name}-execution-role`
- `${data.aws_default_tags.current.tags.environment-name}-execution-role-${data.aws_region.current.name}`
- `batch-manifests-${data.aws_default_tags.current.tags.application}-${data.aws_default_tags.current.tags.account-name}-${data.aws_region.current.name}/*`

IAM policies support wildcards in the resource name, so we can use wildcards to grant access to all resources that match a pattern. This is useful when granting access to resources that are created dynamically.

Adopting a consistent naming convention for resources will make it easier to grant access to resources and avoid conflicts.

## Decision

Use a consistent naming convention for resources in AWS. The naming convention should include the following values in this order:

- `resource-name` describing the role/function of the resource (e.g. `event-received`)
- `application-name` which is the product name (e.g. `opg-modernising-lpa`)
- `account-name` which is the AWS account name (e.g. `development`)
- `region-name` which is the AWS region name (e.g. `eu-west-1`)
- `environment-name` which is the environment name (e.g. `production`)

It isn't necessary to include resource type in the name because the resource type is already specified in the resource definition.

`application-name` and `account-name` will be used when a resource name must be globally unique, such as an S3 bucket name. They can be omitted if the resource is not globally unique.

(this leads to the consequence that application name used in aws_kms_alias is not necessary)

`account-name` should be used for resources shared at account level.

`region-name` only use region if the resource must be globally unique, for example, S3 bucket names and IAM policy names.

`environment-name` should be used for resources that are environment-specific.

(this leads to the consequence that resources will have either `environment-name` or `account-name` and not likely to have both)

## Consequences

- Resources will be easier to identify and grant access to
- some resources will need renaming

resources that must include region name;

- s3 buckets
- IAM policy names
- cloudwatch metric alarms
