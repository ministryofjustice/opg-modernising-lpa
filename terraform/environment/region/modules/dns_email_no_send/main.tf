# internal facing domain SPF and DMARC Records
resource "aws_route53_record" "spf" {
  provider = aws.management
  zone_id  = var.aws_route53_zone.zone_id
  name     = "${var.dns_name}${var.aws_route53_zone.name}"
  type     = "TXT"
  ttl      = "300"

  records = [
    "v=spf1 -all",
  ]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "spf_redirect" {
  count    = data.aws_default_tags.current.tags.environment-name == "production" ? 1 : 0
  provider = aws.management
  zone_id  = var.aws_route53_zone.zone_id
  name     = var.aws_route53_zone.name
  type     = "TXT"
  ttl      = "300"

  records = [
    "v=spf1 -all",
  ]

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_route53_record" "dmarc" {
  provider = aws.management
  zone_id  = var.aws_route53_zone.zone_id
  name     = "_dmarc.${var.dns_name}${var.aws_route53_zone.name}"
  type     = "TXT"
  ttl      = "300"

  records = [
    "v=DMARC1; p=reject; sp=reject; fo=1; rua=mailto:dmarc-rua@dmarc.service.gov.uk; ruf=mailto:dmarc-ruf@dmarc.service.gov.uk",
  ]
}

resource "aws_route53_record" "dmarc_redirect" {
  count    = data.aws_default_tags.current.tags.environment-name == "production" ? 1 : 0
  provider = aws.management
  zone_id  = var.aws_route53_zone.zone_id
  name     = "_dmarc.${var.aws_route53_zone.name}"
  type     = "TXT"
  ttl      = "300"

  records = [
    "v=DMARC1; p=reject; sp=reject; fo=1; rua=mailto:dmarc-rua@dmarc.service.gov.uk; ruf=mailto:dmarc-ruf@dmarc.service.gov.uk",
  ]
}
