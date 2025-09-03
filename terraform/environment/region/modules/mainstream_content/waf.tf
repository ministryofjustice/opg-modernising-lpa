resource "aws_wafv2_web_acl_association" "shared_waf" {
  resource_arn = aws_lb.app.arn
  web_acl_arn  = data.aws_wafv2_web_acl.shared.arn
  provider     = aws.region
}

data "aws_wafv2_web_acl" "shared" {
  name     = "shared-${data.aws_default_tags.current.tags.account-name}-web-acl"
  scope    = "REGIONAL"
  provider = aws.region
}
