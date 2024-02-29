# 3. Use AWS OpenSearch service

Date: 2024-02-28

## Status

Accepted

## Context

We currently use DynamoDB to store data. We have a requirement to present
numbered pages of data to users, which is something we cannot do with DynamoDB
(it supports cursor based pagination only, i.e. "load next page"). We also have
future requirements for filtering and searching on other fields, to produce a
dashboard.

## Decision

Deploying AWS OpenSearch will satisify these use cases, without requiring us to
rewrite all of our data access layer. We will be able to index the data required
in addition to our current store. In particular, we will use AWS OpenSearch
Serverless so we don't have to deal with tuning the setup of search, at least
for the initial period of use.

An alternative suggestion, we won't pursue now, is to use Postgres. This would
allow us to use a single store for all uses. But the negative is that we would
need to change all of our current data layer to use it instead of DynamoDB. It
also would mean making further decisions on data management: how would we
migrate schemas, what indexes would we need to build for our queries, and which
package should we use for data access.

## Consequences

- We will have an additional store of data that needs to be kept in sync with a
subset of our main data.
- We will have an additional cost of running OpenSearch.
- We will be able to satisify any future requirements for searching or filtering.
