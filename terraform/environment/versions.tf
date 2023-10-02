terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.19.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.0.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.5.7"
}
