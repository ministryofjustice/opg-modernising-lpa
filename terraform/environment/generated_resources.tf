# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

# __generated__ by Terraform from "health-checks-demo-environment"

resource "aws_cloudwatch_dashboard" "health_check" {
  provider       = aws.global
  dashboard_body = "{\"widgets\":[]}"
  dashboard_name = "health-checks-${local.environment_name}-environment"
}
