terraform {
  backend "s3" {
    bucket         = "opg.terraform.state"
    key            = "opg-modernising-lpa-environment/terraform.tfstate"
    encrypt        = true
    region         = "eu-west-1"
    role_arn       = "arn:aws:iam::311462405659:role/modernising-lpa-ci"
    dynamodb_table = "remote_lock"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.54.0"
    }
  }
  required_version = ">= 1.2.2"
}

variable "default_role" {
  type    = string
  default = "modernising-lpa-ci"
}

variable "default_web_identity_role" {
  type    = string
  default = "modernising-lpa-github-actions-ci-terraform"
}

variable "ci_cd_web_identity_token" {
  type      = string
  default   = "FAILSAFE"
  sensitive = true
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
  assume_role_with_web_identity {
    role_arn           = "arn:aws:iam::${local.environment.account_id}:role/${var.default_web_identity_role}"
    session_name       = "opg-modernising-lpa-terraform-web-identity-session"
    web_identity_token = var.ci_cd_web_identity_token
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
