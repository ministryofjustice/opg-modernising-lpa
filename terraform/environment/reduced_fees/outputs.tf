output "dynamodb_table" {
  value = aws_dynamodb_table.reduced_fees
}

output "event_bus" {
  value = aws_cloudwatch_event_bus.reduced_fees
}
