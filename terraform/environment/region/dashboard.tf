# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.

# __generated__ by Terraform from "health-checks-demo-environment"

resource "aws_cloudwatch_dashboard" "health_check" {
  provider = aws.global
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          sparkline = true,
          view      = "singleValue",
          metrics = [
            ["AWS/Route53", "HealthCheckPercentageHealthy", "HealthCheckId", aws_route53_health_check.service_health_check.id, { region = "us-east-1" }]
          ],
          region = "us-east-1",
          start  = "-PT8640H",
          end    = "P0D",
          period = 300,
          title  = "service health-check - average uptime of the service over 12 month window"
        }
      }
    ]
  })
  dashboard_name = "health-checks-${data.aws_default_tags.current.tags.environment-name}-environment"
}
