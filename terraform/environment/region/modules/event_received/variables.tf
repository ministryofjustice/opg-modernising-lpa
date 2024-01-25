variable "lambda_function_image_ecr_url" {
  type = string
}

variable "lambda_function_image_tag" {
  type = string
}

variable "event_bus_name" {
  type = string
}

variable "lpas_table" {
  type = object({
    arn  = string
    name = string
  })
}

variable "app_public_url" {
  type = string
}

variable "uploads_bucket" {
  type = any
}

variable "uid_base_url" {
  type = string
}

variable "allowed_api_arns" {
  type = list(string)
}
