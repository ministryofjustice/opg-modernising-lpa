data "aws_default_tags" "current" {
  provider = aws.eu_west_1
}

data "aws_secretsmanager_secret" "opg_metrics_api_key" {
  name     = "opg-metrics-api-key/mrlpa-${data.aws_default_tags.current.tags.account-name}"
  provider = aws.shared_eu_west_1
}

data "aws_secretsmanager_secret_version" "opg_metrics_api_key" {
  secret_id     = data.aws_secretsmanager_secret.opg_metrics_api_key.id
  version_stage = "AWSCURRENT"
  provider      = aws.shared_eu_west_1
}

resource "aws_cloudwatch_event_connection" "opg_metrics" {
  name               = "opg-metrics"
  description        = "Account level - connection and auth for opg-metrics"
  authorization_type = "API_KEY"

  auth_parameters {
    api_key {
      key   = "x-api-key"
      value = data.aws_secretsmanager_secret_version.opg_metrics_api_key.secret_string
    }
  }
  provider = aws.eu_west_1
}

resource "aws_cloudwatch_event_api_destination" "opg_metrics_put" {
  name                             = "opg-metrics metrics PUT"
  description                      = "Account level - an endpoint to push metrics to"
  invocation_endpoint              = "${local.account.opg_metrics_endpoint}/metrics"
  http_method                      = "PUT"
  invocation_rate_limit_per_second = 300
  connection_arn                   = aws_cloudwatch_event_connection.opg_metrics.arn
  provider                         = aws.eu_west_1
}

resource "aws_ssm_parameter" "opg_metrics_arn" {
  name     = "opg-metrics-api-destination-arn"
  type     = "String"
  value    = aws_cloudwatch_event_api_destination.opg_metrics_put.arn
  provider = aws.eu_west_1
}
