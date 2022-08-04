data "aws_route53_zone" "modernising_lpa" {
  provider = aws.management_global
  name     = "modernising.opg.service.justice.gov.uk"
}

locals {
  dns_namespace_for_environment = local.environment_name == "production" ? "" : "${local.environment_name}."
}

resource "aws_route53_record" "app" {
  # app.modernising.opg.service.justice.gov.uk
  provider = aws.management_global
  zone_id  = data.aws_route53_zone.modernising_lpa.zone_id
  name     = "${local.dns_namespace_for_environment}app.${data.aws_route53_zone.modernising_lpa.name}"
  type     = "A"

  alias {
    evaluate_target_health = false
    name                   = module.eu_west_1.app_load_balancer.dns_name
    zone_id                = module.eu_west_1.app_load_balancer.zone_id
  }

  lifecycle {
    create_before_destroy = true
  }
}

output "app_dns" {
  value = aws_route53_record.app.fqdn
}
