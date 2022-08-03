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

resource "aws_iam_role_policy" "execution_role" {
  name     = "${local.environment_name}-execution-role"
  policy   = data.aws_iam_policy_document.execution_role.json
  role     = aws_iam_role.execution_role.id
  provider = aws.global
}

data "aws_iam_policy_document" "execution_role" {
  statement {
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
    ]
  }
  statement {
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
  }
  provider = aws.global
}
