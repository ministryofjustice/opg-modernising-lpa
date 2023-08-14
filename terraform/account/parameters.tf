resource "aws_ssm_parameter" "additional_allowed_ingress" {
  name  = "/modernising-lpa/additional-allowed-ingress-cidrs/${data.aws_default_tags.global.tags.account-name}"
  type  = "StringList"
  value = "[default]"
  lifecycle {
    ignore_changes = [value]
  }
  provider = aws.management_global
}
