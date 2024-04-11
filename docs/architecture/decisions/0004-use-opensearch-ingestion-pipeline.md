# 3. Use AWS OpenSearch Ingestion Pipelines

Date: 2024-04-08

## Status

Accepted

## Context

The app service puts items into the Opensearch Serverless index when a lay or
supporter LPA is created so that they can be searched for later.

If the index were lost, we aren't able to recover the index. This would mean
that the item would not be searchable.

If a request fails, the user gets a 500 error and would need to resubmit form
data to trigger the request again.

## Decision

Use Opensearch Ingestion Pipelines will allow us to process the DynamoDB
stream used to maintain global replica tables, and route specific stream events
to be added to the index.

Using the event stream is more reslilient because events are garunteed, and
items are automatically retried.

Ingestion pipelines can also be used to process a DynamoDB export to S3, which
would be a mechanism to recover the index if it were lost.

## Consequences

- The app service will stop putting items into the index directly
- DynamoDB streeam will be enabled for all environments
- An ingestion pipeline will be responsible for adding items to the index from
  the DynamoDB event stream
- An ingestion pipeline will be responsible for recovering the index from a
  table export if it is lost
