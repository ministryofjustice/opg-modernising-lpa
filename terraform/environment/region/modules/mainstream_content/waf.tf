resource "aws_wafv2_web_acl_association" "main" {
  resource_arn = aws_lb.app.arn
  web_acl_arn  = data.aws_wafv2_web_acl.main.arn
  provider     = aws.region
}

data "aws_wafv2_web_acl" "main" {
  provider = aws.region
  name     = "${data.aws_default_tags.current.tags.account-name}-web-acl"
  scope    = "REGIONAL"
}
