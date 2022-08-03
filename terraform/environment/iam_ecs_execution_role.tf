resource "aws_iam_role" "execution_role" {
  name               = "${local.environment_name}-execution-role-ecs-cluster"
  assume_role_policy = data.aws_iam_policy_document.execution_role_assume_policy.json
  provider           = aws.global
}

data "aws_iam_policy_document" "execution_role_assume_policy" {
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
