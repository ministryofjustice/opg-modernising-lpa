# Disabling DynamoDB Global Tables procedure

At present, when replicas are used to create DynamoDB global tables, the AWS API also enables DynamoDB Streams to send on write requests on to the replica table.

However, when disabling replicas, an attempt is made first to disable streams which fails because they are required got Global Tables.

See (https://github.com/hashicorp/terraform-provider-aws/issues/19342)

This is the workaround to use for this issue.

## Enabling Global Tables

When enabling replicas for an environment, update the both `region_replica_enabled` and `stream_enabled` values in the terraform.tfvars.json file to `true`.

```json
"dynamodb": {
  "region_replica_enabled": true,
  "stream_enabled": true
},
```

## Enabling Global Tables

When disbaling Global Tables, set `region_replica_enabled` to `false` first, apply the changes for the environment, then set `stream_enabled` to `false` and apply again.
