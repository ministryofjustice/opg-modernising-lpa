# Parameters locations in SSM Parameter Store

Date: 2024-06-19

## Status

Accepted

## Context

Over time, AWS SSM parameters have been created in multiple locations to make them available for use across regions, accounts and services. Knowing where to go to access or if required modify parameter values isn't obvious.

Presently we have parameters in the following locations;

### Managed by MRLPA

Management Account - us-east-1 region

```text
/modernising-lpa/container-version/${local.environment_name}
/modernising-lpa/dns-target-region/${local.environment_name}
/modernising-lpa/additional-allowed-ingress-cidrs/${data.aws_default_tags.global.tags.account-name}
```

MRLPA Account (development, preproduction or production) - deployed region (eg. eu-west-1, eu-west-2)

```text
/modernising-lpa/s3-batch-configuration/${data.aws_default_tags.current.tags.environment-name}/s3_batch_configuration
```

### Used by MRLPA - Managed by Sirius

Management Account - deployed region (eg. eu-west-1, eu-west-2)

```text
/modernising-lpa/reduced_fees_uploads_bucket_kms_key_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.region}
/modernising-lpa/reduced_fees_uploads_bucket_arn/${var.reduced_fees.target_environment}/${data.aws_region.current.region}
```

We can make it easier to work with parameters by rationalising the locations and defining how we choose where to locate them.

## Decision

- Use global region (us-east-1) for parameters used cross regions (eg. container version used in eu-west-1 and eu-west-2)
- Use management account for parameters used across services (eg. Sirius, github actions)
- Use local account (developement, preproduction, production) and local region for other use cases

## Consequences
