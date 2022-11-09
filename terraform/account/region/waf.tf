resource "aws_wafv2_web_acl" "main" {
  provider    = aws.region
  name        = "${data.aws_default_tags.current.tags.account-name}-web-acl"
  description = "Managed rules"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${data.aws_default_tags.current.tags.account-name}-web-acl"
    sampled_requests_enabled   = true
  }
}

resource "aws_wafv2_web_acl_logging_configuration" "main" {
  provider                = aws.region
  log_destination_configs = [aws_cloudwatch_log_group.waf_web_acl.arn]
  resource_arn            = aws_wafv2_web_acl.main.arn
}

resource "aws_cloudwatch_log_group" "waf_web_acl" {
  provider          = aws.region
  name              = "aws-waf-logs-${data.aws_default_tags.current.tags.account-name}"
  retention_in_days = 120
  kms_key_id        = data.aws_kms_alias.cloudwatch_application_logs_encryption.target_key_arn
  tags = {
    "Name" = "${data.aws_default_tags.current.tags.account-name}-web-acl"
  }
}
