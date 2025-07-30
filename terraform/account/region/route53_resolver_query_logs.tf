resource "aws_cloudwatch_log_group" "main" {
  name     = "${data.aws_default_tags.current.tags.account-name}-route53-resolver-logs-${data.aws_region.current.name}"
  provider = aws.region
}

resource "aws_route53_resolver_query_log_config" "main" {
  name            = "main"
  destination_arn = aws_cloudwatch_log_group.main.arn
}

resource "aws_route53_resolver_query_log_config_association" "main" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.main.id
  resource_id                  = module.network.vpc.id
  provider                     = aws.region
}
