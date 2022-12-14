resource "aws_backup_vault" "eu_west_1" {
  name     = "eu-west-1-${local.default_tags.account-name}-backup-vault"
  provider = aws.eu_west_1
}

resource "aws_backup_vault" "eu_west_2" {
  name     = "eu-west-2-${local.default_tags.account-name}-backup-vault"
  provider = aws.eu_west_2
}
