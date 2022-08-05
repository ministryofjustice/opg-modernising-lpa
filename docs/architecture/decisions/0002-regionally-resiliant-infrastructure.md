# 2. Regionally resiliant infrastructure

Date: 2022-08-05

## Status

Accepted

## Context

The issue motivating this decision, and any context that influences or constrains the decision.

Disaster recovery and global high availability are difficult to introduce to infrastructure later, and slow to achieve when not considered first.

We know already that the Modernising LPA service will need to meet short RTO and RPO objectives.

## Decision

We will design infrastructure in a way that enables disaster recovery and global high availability strategies.

Infrastructure will be organsised in terraform as regional and global at a per-account and per-environment level.

Modules will be used to define a region Resources within a region will be defined as sub-modules also.

## Consequences

Infrastructure will be easier to implement in a way that enables disaster recovery planning and high availability with RTO and RPO times counted in minutes

Infrastructure will be easy to replcicate across regions, with shared global resources between each region.

For example the account terraform configuration will have a structure like this

```shell
.
├── region
│   ├── modules
│   │   └── certificates
│   │       ├── main.tf
│   │       └── terraform.tf
│   ├── certificates.tf
│   ├── network.tf
│   ├── terraform.tf
│   └── variables.tf
├── README.md
├── regions.tf
├── kms.tf
├── terraform.tf
```

Regions.tf will instatiate the /region module for each AWS region required.

Rsources inside /region will be grouped as modules also, allowing for parts of a region to be replicated as and when needed.

This will allow us to deploy the service in a way that is globally resiliant, and highly available.

Global resources such as IAM roles, or Route53 records will be created in a global region.

Where a resource supports regional replication such as DynamoDB tables, of KMS keys, they will exist in the global layer.
