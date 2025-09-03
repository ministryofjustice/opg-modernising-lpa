resource "aws_iam_role" "mainstream_content_task_role" {
  name               = "${data.aws_default_tags.current.tags.environment-name}-mrlpa-mc-app-task-role"
  assume_role_policy = data.aws_iam_policy_document.task_role_assume_policy.json
  provider           = aws.global
}
