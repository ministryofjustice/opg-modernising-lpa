data "aws_route53_zone" "modernising_lpa" {
  provider = aws.management_global
  name     = "modernising.opg.service.justice.gov.uk"
}

locals {
  dns_namespace_for_environment               = data.aws_default_tags.current.tags.environment-name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}."
  dns_namespace_for_environment_mock_onelogin = data.aws_default_tags.current.tags.environment-name == "production" ? "" : "${data.aws_default_tags.current.tags.environment-name}-mock-onelogin."
}

resource "aws_route53_record" "app" {
  # app.modernising.opg.service.justice.gov.uk
  provider       = aws.management_global
  zone_id        = data.aws_route53_zone.modernising_lpa.zone_id
  name           = "${local.dns_namespace_for_environment}app.${data.aws_route53_zone.modernising_lpa.name}"
  type           = "A"
  set_identifier = data.aws_region.current.name

  alias {
    evaluate_target_health = false
    name                   = module.app.load_balancer.dns_name
    zone_id                = module.app.load_balancer.zone_id
  }

  weighted_routing_policy {
    weight = var.dns_weighting
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "mock_onelogin" {
  # *-mock-onelogin.app.modernising.opg.service.justice.gov.uk
  count          = length(module.mock_onelogin)
  provider       = aws.management_global
  zone_id        = data.aws_route53_zone.modernising_lpa.zone_id
  name           = "${local.dns_namespace_for_environment_mock_onelogin}app.${data.aws_route53_zone.modernising_lpa.name}"
  type           = "A"
  set_identifier = data.aws_region.current.name

  alias {
    evaluate_target_health = false
    name                   = module.mock_onelogin[0].load_balancer.dns_name
    zone_id                = module.mock_onelogin[0].load_balancer.zone_id
  }

  weighted_routing_policy {
    weight = var.dns_weighting
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_service_discovery_private_dns_namespace" "internal" {
  name        = "${data.aws_default_tags.current.tags.environment-name}.internal.modernising.ecs"
  description = "Private DNS namespace modernising services"
  vpc         = data.aws_vpc.main.id
  provider    = aws.region
}
