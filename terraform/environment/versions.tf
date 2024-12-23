terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.82.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.18.2"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.10.3"
}
