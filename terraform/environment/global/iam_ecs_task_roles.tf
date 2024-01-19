resource "aws_iam_role" "app_task_role" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-app-task-role"
  assume_role_policy = data.aws_iam_policy_document.task_role_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "task_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}

