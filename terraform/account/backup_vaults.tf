module "aws_backup_vaults_eu_west_1" {
  source = "./modules/aws_backup_vault"
  providers = {
    aws.region = aws.eu_west_1
    aws.global = aws.global
  }
}

module "aws_backup_vaults_eu_west_2" {
  source = "./modules/aws_backup_vault"
  providers = {
    aws.region = aws.eu_west_2
    aws.global = aws.global
  }
}
