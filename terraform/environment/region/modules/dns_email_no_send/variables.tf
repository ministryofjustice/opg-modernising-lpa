variable "aws_route53_zone" {
  type = object({
    zone_id = string
    name    = string
  })
  description = "A Route53 Hosted zone_id and name"
}

variable "dns_name" {
  type        = string
  description = "prefix part of the service url for example $${local.dns_namespace_for_environment}app."
}
