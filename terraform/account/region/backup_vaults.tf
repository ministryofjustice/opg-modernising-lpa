module "aws_backup_vaults" {
  source = "./modules/aws_backup_vault"
  providers = {
    aws.region = aws.region
    aws.global = aws.global
  }
}
