module "dns_firewall" {
  source                             = "./modules/dns_firewall"
  vpc_id                             = module.network.vpc.id
  cloudwatch_log_group_kms_key_alias = var.cloudwatch_log_group_kms_key_alias
  providers = {
    aws.region = aws.region
  }
}
