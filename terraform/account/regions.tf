module "eu_west_1" {
  source             = "./region"
  network_cidr_block = "10.162.0.0/16"
  providers = {
    aws.region     = aws.eu_west_1
    aws.management = aws.management_eu_west_1
  }

}
