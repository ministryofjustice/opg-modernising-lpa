resource "aws_iam_role" "schedule_runner_scheduler" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-scheduler-role"
  assume_role_policy = data.aws_iam_policy_document.schedule_runner_scheduler_assume.json
  provider           = aws.global
}

data "aws_iam_policy_document" "schedule_runner_scheduler_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["scheduler.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}
