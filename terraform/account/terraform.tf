terraform {
  backend "s3" {
    bucket         = "opg.terraform.state"
    key            = "opg-modernising-lpa-account/terraform.tfstate"
    encrypt        = true
    region         = "eu-west-1"
    role_arn       = "arn:aws:iam::311462405659:role/modernising-lpa-ci"
    dynamodb_table = "remote_lock"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.49.0"
    }
  }
  required_version = ">= 1.2.6"
}

variable "default_role" {
  type    = string
  default = "modernising-lpa-ci"
}
variable "management_role" {
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
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
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
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
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
  alias  = "global"
  region = "us-east-1"
  default_tags {
    tags = local.default_tags
  }
  assume_role {
    role_arn     = "arn:aws:iam::${local.account.account_id}:role/${var.default_role}"
    session_name = "opg-modernising-lpa-terraform-session"
  }
}

data "aws_region" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_caller_identity" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_default_tags" "eu_west_1" {
  provider = aws.eu_west_1
}

data "aws_region" "eu_west_2" {
  provider = aws.eu_west_2
}

data "aws_caller_identity" "eu_west_2" {
  provider = aws.eu_west_2
}

data "aws_default_tags" "eu_west_2" {
  provider = aws.eu_west_2
}

data "aws_region" "global" {
  provider = aws.global
}

data "aws_caller_identity" "global" {
  provider = aws.global
}

data "aws_default_tags" "global" {
  provider = aws.global
}
