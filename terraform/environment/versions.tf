terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.32.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.5.1"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.6.6"
}
