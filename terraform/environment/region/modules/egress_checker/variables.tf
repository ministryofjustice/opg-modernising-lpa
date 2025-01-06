variable "lambda_function_image_ecr_url" {
  type = string
}

variable "lambda_function_image_tag" {
  type = string
}

variable "egress_checker_lambda_role" {
  type = any
}

variable "vpc_config" {
  description = "Configuration block for VPC"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
}
