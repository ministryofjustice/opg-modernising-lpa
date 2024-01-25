terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.33.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.4.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.6.6"
}
