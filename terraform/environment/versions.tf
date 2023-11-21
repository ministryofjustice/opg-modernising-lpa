terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.26.0"
    }
    pagerduty = {
      source  = "PagerDuty/pagerduty"
      version = "~> 3.1.0"
    }
    local = {
      source = "hashicorp/local"
    }
  }
  required_version = "1.6.4"
}
