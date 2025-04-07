terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.94.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.23.1"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.11.3"
}
