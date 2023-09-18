terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.17.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 2.16.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.5.7"
}
