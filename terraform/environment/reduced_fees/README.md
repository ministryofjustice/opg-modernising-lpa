# Reduced Fees events

This module creates event driven architecture to send and receive reduced fees events between MLPAB and Sirius.

Outbound events are sent to an Event Bridge event bus from DynamoDB Streams using EventBridge Pipes.

From there the event is sent to the Sirius event bus for processing using a rule with cross account putEvent IAM permissions.

You can create an incoming reduced fees event by using the following aws cli put-events command:

```shell
aws-vault exec mlpa-dev -- aws events put-events --entries file://reduced_fees_update_event.json
```
