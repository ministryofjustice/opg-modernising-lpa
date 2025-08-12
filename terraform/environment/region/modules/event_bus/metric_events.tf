resource "aws_cloudwatch_event_rule" "metric_events" {
  name           = "${data.aws_default_tags.current.tags.environment-name}-metric-events"
  description    = "forward events to opg-metrics service"
  event_bus_name = aws_cloudwatch_event_bus.main.name

  event_pattern = jsonencode({
    source      = ["opg.poas.makeregister"]
    detail-type = ["metric"]
  })
  provider = aws.region
}

data "aws_ssm_parameter" "opg_metrics_arn" {
  name     = "opg-metrics-api-destination-arn"
  provider = aws.region
}

resource "aws_iam_role_policy" "opg_metrics" {
  name     = "opg-metrics-${data.aws_region.current.region}"
  role     = var.opg_metrics_api_destination_role.name
  policy   = data.aws_iam_policy_document.opg_metrics.json
  provider = aws.region
}


data "aws_iam_policy_document" "opg_metrics" {
  statement {
    effect  = "Allow"
    actions = ["events:InvokeApiDestination"]
    resources = [
      "${data.aws_ssm_parameter.opg_metrics_arn.value}*"
    ]
  }
  provider = aws.global
}

resource "aws_cloudwatch_event_target" "opg_metrics" {
  arn            = data.aws_ssm_parameter.opg_metrics_arn.insecure_value
  event_bus_name = aws_cloudwatch_event_bus.main.name
  rule           = aws_cloudwatch_event_rule.metric_events.name
  role_arn       = var.opg_metrics_api_destination_role.arn
  http_target {
    header_parameters = {
      Content-Type = "application/json"
    }
    path_parameter_values   = []
    query_string_parameters = {}
  }
  input_path = "$.detail"
  provider   = aws.region
}

resource "aws_cloudwatch_event_target" "metrics_emitted" {
  count          = var.log_emitted_events ? 1 : 0
  target_id      = "${data.aws_default_tags.current.tags.environment-name}-metrics-emitted-target"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  rule           = aws_cloudwatch_event_rule.metric_events.name
  arn            = aws_cloudwatch_log_group.events_emitted[0].arn
  provider       = aws.region
}
