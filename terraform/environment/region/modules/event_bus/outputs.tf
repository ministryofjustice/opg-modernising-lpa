output "event_bus" {
  value = aws_cloudwatch_event_bus.main
}

output "event_bus_dead_letter_queue" {
  value = aws_sqs_queue.event_bus_dead_letter_queue
}
