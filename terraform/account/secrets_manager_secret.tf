resource "aws_secretsmanager_secret" "private_jwt_key_base64" {
  name       = "private-jwt-key-base64"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "public_jwt_key_base64" {
  name       = "public-jwt-key-base64"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}
