terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.44.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.11.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.7.5"
}
