terraform {
  backend "s3" {
    bucket         = "opg.terraform.state"
    key            = "opg-modernising-lpa-environment/terraform.tfstate"
    encrypt        = true
    region         = "eu-west-1"
    role_arn       = "arn:aws:iam::311462405659:role/modernising-lpa-state-access"
    dynamodb_table = "remote_lock"
  }
}

variable "default_role" {
  type    = string
  default = "modernising-lpa-ci"
}

provider "aws" {
  alias  = "eu_west_1"
  region = "eu-west-1"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::${local.environment.account_id}:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "aws" {
  alias  = "eu_west_2"
  region = "eu-west-2"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::${local.environment.account_id}:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "aws" {
  alias  = "global"
  region = "us-east-1"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::${local.environment.account_id}:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "aws" {
  alias  = "management_global"
  region = "us-east-1"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::311462405659:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "aws" {
  alias  = "management_eu_west_2"
  region = "eu-west-2"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::311462405659:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "aws" {
  alias  = "management_eu_west_1"
  region = "eu-west-1"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::311462405659:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

provider "pagerduty" {
  token = var.pagerduty_api_key
}
