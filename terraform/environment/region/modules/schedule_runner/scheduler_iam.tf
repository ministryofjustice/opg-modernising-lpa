data "aws_iam_policy_document" "lambda_access_policy" {
  statement {
    sid       = "allowLambdaInvoke"
    effect    = "Allow"
    resources = [module.schedule_runner.lambda.arn]
    actions = [
      "lambda:Invoke",
    ]
  }
  provider = aws.region
}

resource "aws_iam_role_policy" "lambda_access_role_policy" {
  policy   = data.aws_iam_policy_document.lambda_access_policy.json
  role     = var.schedule_runner_scheduler.arn
  provider = aws.region
}
