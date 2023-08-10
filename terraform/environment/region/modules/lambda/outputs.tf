output "lambda" {
  description = "The lambda function"
  value       = aws_lambda_function.lambda_function
}

output "lambda_log" {
  description = "The lambda logs"
  value       = aws_cloudwatch_log_group.lambda
}

output "lambda_role" {
  description = "The lambda role"
  value       = aws_iam_role.lambda_role
}
