data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

data "pagerduty_service" "main" {
  name = local.account.pagerduty_service_name
}

resource "pagerduty_service_integration" "main" {
  name    = "Modernising LPA ${data.pagerduty_vendor.cloudwatch.name} Generic Notification"
  service = data.pagerduty_service.main.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}
