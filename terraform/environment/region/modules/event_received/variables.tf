variable "lambda_function_image_ecr_arn" {
  type = string
}

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
