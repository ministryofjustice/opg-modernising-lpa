terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.81.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "3.18.1"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.10.2"
}
