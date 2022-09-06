resource "aws_backup_vault" "main" {
  name = "${data.aws_region.current.name}-${data.aws_default_tags.current.tags.environment-name}-backup-vault"
}
