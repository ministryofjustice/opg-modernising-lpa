resource "aws_iam_role" "grafana_cross_account_reader_role" {
  name = "grafana-cross-account-log-reader"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::679638075911:root"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "grafana_cw_logs_read_only" {
  provider   = aws.region
  role       = aws_iam_role.grafana_cross_account_reader_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsReadOnly"
}
