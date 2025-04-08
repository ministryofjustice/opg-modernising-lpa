resource "aws_cloudwatch_event_rule" "metric_events" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-metric-events"
  description    = "forward events to opg-metrics service"
  event_bus_name = aws_cloudwatch_event_bus.main.name

  event_pattern = jsonencode({
    source      = ["opg.poas.makeregister"]
    detail-type = ["metric*"]
  })
  provider = aws.region
}

data "aws_ssm_parameter" "opg_metrics_arn" {
  name     = "opg-metrics-api-destination-arn"
  provider = aws.region
}

resource "aws_cloudwatch_event_target" "opg_metrics" {
  arn            = data.aws_ssm_parameter.opg_metrics_arn.insecure_value
  event_bus_name = aws_cloudwatch_event_bus.main.name
  rule           = aws_cloudwatch_event_rule.metric_events.name
  # role_arn       = "arn:aws:iam::653761790766:role/service-role/Amazon_EventBridge_Invoke_Api_Destination_344039630"
  role_arn = var.opg_metrics_api_destination_role.arn
  http_target {
    header_parameters       = {}
    path_parameter_values   = []
    query_string_parameters = {}
  }
  provider = aws.region
}
