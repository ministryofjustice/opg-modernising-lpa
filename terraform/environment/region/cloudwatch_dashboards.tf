locals {
  security_template_vars = {
    region                = data.aws_region.current.name
    account_name          = data.aws_default_tags.current.tags.account-name
    environment_name      = data.aws_default_tags.current.tags.environment-name
    app_loadbalancer_name = module.app.load_balancer.arn_suffix
    nat_gateway_a         = data.aws_nat_gateway.main[0].id
    nat_gateway_b         = data.aws_nat_gateway.main[1].id
    nat_gateway_c         = data.aws_nat_gateway.main[2].id
  }
}

resource "aws_cloudwatch_dashboard" "security" {
  provider       = aws.region
  dashboard_name = "${data.aws_default_tags.current.tags.environment-name}-Security"
  dashboard_body = templatefile(
    "cloudwatch_dashboards/security.json.tftpl",
    local.security_template_vars
  )
}

resource "aws_cloudwatch_dashboard" "health_checks" {
  dashboard_name = "${data.aws_default_tags.current.tags.environment-name}-Health"
  provider       = aws.region
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
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 5

        properties = {
          sparkline = true,
          view      = "singleValue",
          metrics = [
            ["AWS/Route53", "HealthCheckPercentageHealthy", "HealthCheckId", aws_route53_health_check.dependency_health_check.id, { region = "us-east-1" }]
          ],
          region = "us-east-1",
          start  = "-PT8640H",
          end    = "P0D",
          period = 300,
          title  = "dependency health-check - average availability of service dependencies over 12 month window"
        }
      }
    ]
  })
}
