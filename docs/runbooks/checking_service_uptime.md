# Checking service uptime

## Overview

This runbook describes how to check the uptime of the service, and how to check the uptime of the service's dependencies.

Health checks are defined in the [adr-007](https://docs.opg.service.justice.gov.uk/documentation/adrs/adr-007.html) ADR.

We have metrics for the `/health-check/service` endpoint and the `/health-check/dependencies` endpoint.

Both endpoints are monitored by a Route53 health check that runs every 30 seconds. The health check is configured to send a notification to the team via Slack if the endpoint is down.

The [Route53 Health checks](https://us-east-1.console.aws.amazon.com/route53/healthchecks/home?region=us-east-1#/) are in the AWS us-east-1 region, and check from locations in the US, EU and Asia.

## Checking the uptime of the service

Each environment has a Cloudwatch dashboard that shows the uptime of the service and it's dependencies, named `health-checks-<environment-name>-environment`.

You can access them here:

- [Cloudwatch Dashboards](https://eu-west-1.console.aws.amazon.com/cloudwatch/home?region=eu-west-1#dashboards)
