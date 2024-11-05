output "ecs_autoscaling_alarm_sns_topic" {
  value = aws_sns_topic.ecs_autoscaling_alarms
}

output "opensearch_lpas_collection_vpc_endpoint" {
  value = aws_opensearchserverless_vpc_endpoint.lpas_collection_vpc_endpoint
}
