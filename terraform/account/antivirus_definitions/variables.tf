variable "account_name" {
  description = "AWS account name"
}

variable "ecr_image_uri" {
  description = "URI of ECR image to use for Lambda"
}

locals {
  environment = split("_", terraform.workspace)[0]
}
