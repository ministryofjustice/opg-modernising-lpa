terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.12.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 2.15.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "= 1.5.2"
}
