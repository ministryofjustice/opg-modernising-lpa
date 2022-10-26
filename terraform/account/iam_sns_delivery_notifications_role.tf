resource "aws_iam_role" "sns_success_feedback" {
  provider           = aws.global
  name               = "sns-success-feedback-role"
  assume_role_policy = data.aws_iam_policy_document.sns_feedback_assume_policy.json
}

resource "aws_iam_role" "sns_failure_feedback" {
  provider           = aws.global
  name               = "sns-failure-feedback-role"
  assume_role_policy = data.aws_iam_policy_document.sns_feedback_assume_policy.json
}

resource "aws_iam_role_policy" "sns_success_feedback" {
  provider = aws.global
  name     = "sns-delivery-notifcations-role"
  policy   = data.aws_iam_policy_document.sns_feedback_actions.json
  role     = aws_iam_role.sns_success_feedback.id
}

resource "aws_iam_role_policy" "sns_failure_feedback" {
  provider = aws.global
  name     = "sns-delivery-notifcations-role"
  policy   = data.aws_iam_policy_document.sns_feedback_actions.json
  role     = aws_iam_role.sns_failure_feedback.id
}

data "aws_iam_policy_document" "sns_feedback_assume_policy" {
  provider = aws.global
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["ecs-tasks.amazonaws.com"]
      type        = "Service"
    }
  }
}

data "aws_iam_policy_document" "sns_feedback_actions" {
  provider = aws.global
  statement {
    effect    = "Allow"
    resources = ["*"] #tfsec:ignore:aws-iam-no-policy-wildcards
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:PutMetricFilter",
      "logs:PutRetentionPolicy"
    ]
  }
}
