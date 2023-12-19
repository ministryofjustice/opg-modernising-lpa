resource "aws_iam_role" "fault_injection_simulator" {
  name               = "fault-injection-simulator-${data.aws_default_tags.current.tags.environment-name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
  provider           = aws.global
}

data "aws_iam_policy_document" "fault_injection_simulator_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["fis.amazonaws.com"]
    }
  }
  provider = aws.global
}

resource "aws_iam_role_policy_attachment" "fault_injection_simulator" {
  role       = aws_iam_role.fault_injection_simulator.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSFaultInjectionSimulatorECSAccess"
  provider   = aws.global
}
