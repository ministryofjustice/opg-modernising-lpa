data "" "main" {
  provider = aws.region
  name     = "${data.aws_default_tags.current.tags.account-name}-web-acl"
  scope    = "REGIONAL"
}

resource "aws_wafv2_web_acl_association" "app" {
  provider     = aws.region
  resource_arn = aws_lb.app.arn
  web_acl_arn  = data.aws_wafv2_web_acl.main.arn
}
