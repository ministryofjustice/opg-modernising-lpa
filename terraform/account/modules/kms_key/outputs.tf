output "kms_key_alias_name" {
  value = aws_kms_alias.main_eu_west_1.name
}

output "kms_target_key_arn_eu_west_1" {
  value = aws_kms_alias.main_eu_west_1.target_key_arn
}

output "kms_target_key_arn_eu_west_2" {
  value = aws_kms_alias.main_eu_west_2.target_key_arn
}
