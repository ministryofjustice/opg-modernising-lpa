terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.89.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.21.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.11.0"
}
