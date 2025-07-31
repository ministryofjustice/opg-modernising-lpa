#tfsec:ignore:aws-cloudwatch-log-group-customer-key:exp:2025-08-30
resource "aws_cloudwatch_log_group" "route_53_resolver_logs" {
  name              = "${data.aws_default_tags.current.tags.account-name}-route53-resolver-logs-${data.aws_region.current.name}"
  retention_in_days = 400
  provider          = aws.region
}

resource "aws_route53_resolver_query_log_config" "main" {
  name            = "main"
  destination_arn = aws_cloudwatch_log_group.route_53_resolver_logs.arn
  provider        = aws.region
}

resource "aws_route53_resolver_query_log_config_association" "main" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.main.id
  resource_id                  = module.network.vpc.id
  provider                     = aws.region
}
