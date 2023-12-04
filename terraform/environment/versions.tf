terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.29.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.2.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.6.5"
}
