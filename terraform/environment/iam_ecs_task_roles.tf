resource "aws_iam_role" "app_task_role" {
  name               = "${local.environment_name}-app-task-role"
  assume_role_policy = data.aws_iam_policy_document.task_role_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "task_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = [
      "sts:AssumeRole",
      "secretsmanager:GetSecretValue",
    ]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}
