resource "aws_cloudwatch_log_group" "aws_route53_resolver_query_log" {
  name              = "route53-resolver-query-log"
  retention_in_days = 400
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  tags = {
    "Name" = "route53-resolver-query-log"
  }
  provider = aws.region
}

resource "aws_route53_resolver_query_log_config" "egress" {
  name            = "egress"
  destination_arn = aws_cloudwatch_log_group.aws_route53_resolver_query_log.arn
  provider        = aws.region
}

resource "aws_route53_resolver_query_log_config_association" "egress" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.egress.id
  resource_id                  = var.vpc_id
  provider                     = aws.region
}

locals {
  service_id = [
    "dynamodb",
    "ecr.api",
    "ecr",
    "events",
    "kms",
    "logs",
    "s3",
    "secretsmanager",
    "xray",
  ]
}

data "aws_service" "services" {
  for_each   = toset(local.service_id)
  region     = data.aws_region.current.name
  service_id = each.value
  provider   = aws.region
}

locals {
  aws_service_dns_name = [for service in data.aws_service.services : "${service.dns_name}."]
  interpolated_dns = [
    "311462405659.dkr.ecr.${data.aws_region.current.name}.amazonaws.com.",
    "prod-${data.aws_region.current.name}-starport-layer-bucket.s3.${data.aws_region.current.name}.amazonaws.com.",
    "public-keys.auth.elb.${data.aws_region.current.name}.amazonaws.com.",
    "public.ecr.aws.",
  ]
  endpoints_dns = [
    "api.notifications.service.gov.uk.",
    "api.os.uk.",
    "current.cvd.clamav.net.",
    "database.clamav.net.",
    "development.lpa-uid.api.opg.service.justice.gov.uk.",
    "integration.lpa-uid.api.opg.service.justice.gov.uk.",
    "oidc.integration.account.gov.uk.",
    "publicapi.payments.service.gov.uk.",
  ]
}
resource "aws_route53_resolver_firewall_domain_list" "egress_allow" {
  name = "egress_allowed"
  domains = concat(
    local.interpolated_dns,
    local.aws_service_dns_name,
    local.endpoints_dns,
  )
  provider = aws.region
}

resource "aws_route53_resolver_firewall_domain_list" "egress_block" {
  name     = "egress_blocked"
  domains  = ["*."]
  provider = aws.region
}

resource "aws_route53_resolver_firewall_rule_group" "egress" {
  name     = "egress"
  provider = aws.region
}

resource "aws_route53_resolver_firewall_rule" "egress_allow" {
  name                    = "egress_allowed"
  action                  = "ALLOW"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.egress_allow.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.egress.id
  priority                = 1
  provider                = aws.region
}

resource "aws_route53_resolver_firewall_rule" "egress_block" {
  name   = "egress_blocked"
  action = "ALERT"
  # action                  = "BLOCK"
  # block_response          = "NODATA"
  firewall_domain_list_id = aws_route53_resolver_firewall_domain_list.egress_block.id
  firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.egress.id
  priority                = 2
  provider                = aws.region
}

resource "aws_route53_resolver_firewall_rule_group_association" "egress" {
  name                   = "egress"
  firewall_rule_group_id = aws_route53_resolver_firewall_rule_group.egress.id
  priority               = 101
  vpc_id                 = var.vpc_id
  provider               = aws.region
}


resource "aws_cloudwatch_query_definition" "dns_firewall_statistics" {
  name = "DNS Firewall Queries/DNS Firewall Statistics"

  log_group_names = [aws_cloudwatch_log_group.aws_route53_resolver_query_log.name]

  query_string = <<EOF
fields @timestamp, query_name, firewall_rule_action
| sort @timestamp desc
| stats count() as frequency by query_name, firewall_rule_action
EOF
  provider     = aws.region
}
