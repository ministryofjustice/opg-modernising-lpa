# Deploying into additional regions

The infrastructure for this project is structured to allow for regional resources such as networks, certificates, and appplication deployment to be provisioned in additional regions.

Other resources such as backup vaults, DNS records and databases are treated as "single entity" global resources with configuration for multi-region replication.

This is so the modernising-lpa service can be made highly available (deployed in and served from multiple places) or to enable a disaster recovery plan with low recovery time and point objectives.

## Deploying into London (eu-west-2)

The terraform configuration includes a definition for a London region that can be "activated". For environment activation, dynamodb table replication must be enabled first.

Raise a pull request to set `region_replica_enabled` and `stream_enabled` to `true` if not already set to true.

To activate and provision account resources into a second region, add `eu-west-2` to the `regions` variable list in `terraform/account/terraform.tfvars.json` for each account.

```json
"regions": [
  "eu-west-1",
  "eu-west-2"
]
```

To activate and provision environment resources into a second region, add `eu-west-2` to the `regions` variable list in `terraform/environment/terraform.tfvars.json` for each account.

```json
"regions": [
  "eu-west-1",
  "eu-west-2"
]
```


Apply this change using an approved pull request.

To deploy

## Deploying into other additional regions

How to add an additional region
