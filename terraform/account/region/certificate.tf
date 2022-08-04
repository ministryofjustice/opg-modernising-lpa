data "aws_route53_zone" "modernising_lpa" {
  provider = aws.management
  name     = "modernising.opg.service.justice.gov.uk"
}

locals {
  dev_wildcard = data.aws_default_tags.current.tags.environment-name == "production" ? "" : "*."
}

resource "aws_acm_certificate" "app" {
  domain_name       = "${local.dev_wildcard}app.modernising.opg.service.justice.gov.uk"
  validation_method = "DNS"
  provider          = aws.region
}

resource "aws_acm_certificate_validation" "app" {
  certificate_arn         = aws_acm_certificate.app.arn
  validation_record_fqdns = [for record in aws_route53_record.certificate_validation_app : record.fqdn]
  provider                = aws.region
}

resource "aws_route53_record" "certificate_validation_app" {
  provider = aws.management
  for_each = {
    for dvo in aws_acm_certificate.app.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.modernising_lpa.zone_id
}
