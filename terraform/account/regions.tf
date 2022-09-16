module "eu_west_1" {
  source             = "./region"
  count              = contains(local.account.regions, "eu-west-1") ? 1 : 0
  network_cidr_block = "10.162.0.0/16"
  providers = {
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
  }
}

module "eu_west_2" {
  source             = "./region"
  count              = contains(local.account.regions, "eu-west-2") ? 1 : 0
  network_cidr_block = "10.162.0.0/16"
  providers = {
    aws.region     = aws.eu_west_2
    aws.management = aws.management_eu_west_2
  }
}

moved {
  from = module.eu_west_1
  to   = module.eu_west_1[0]
}
