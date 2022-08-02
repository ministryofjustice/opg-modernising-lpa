module "network" {
  source                         = "github.com/ministryofjustice/opg-terraform-aws-network?ref=v1.0.0"
  cidr                           = var.network_cidr_block
  default_security_group_ingress = [{}]
  default_security_group_egress  = [{}]
  providers = {
    aws = aws.region
  }
}
