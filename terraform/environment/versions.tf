terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.48.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.11.4"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.8.3"
}
