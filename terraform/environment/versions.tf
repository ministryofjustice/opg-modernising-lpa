terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.49.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.12.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.8.3"
}
