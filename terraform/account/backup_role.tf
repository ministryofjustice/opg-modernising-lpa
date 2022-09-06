resource "aws_iam_role" "aws_backup_role" {
  name               = "aws-backup-role"
  assume_role_policy = data.aws_iam_policy_document.aws_backup_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "aws_backup_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["backup.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}

resource "aws_iam_role_policy_attachment" "aws_backup_role" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
  role       = aws_iam_role.aws_backup_role.name
  provider   = aws.global
}
