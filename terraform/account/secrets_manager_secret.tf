resource "aws_secretsmanager_secret" "private_jwt_key_base64" {
  name       = "private-jwt-key-base64"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "os_postcode_lookup_api_key" {
  name       = "os-postcode-lookup-api-key"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "cookie_session_keys" {
  name       = "cookie-session-keys"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "gov_uk_pay_api_key" {
  name       = "gov-uk-pay-api-key"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "yoti_private_key" {
  name       = "yoti-private-key"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "gov_uk_notify_api_key" {
  name       = "gov-uk-notify-api-key"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "rum_monitor_identity_pool_id" {
  name       = "rum-monitor-identity-pool-id"
  kms_key_id = aws_kms_key.secrets_manager.key_id
  replica {
    kms_key_id = aws_kms_replica_key.secrets_manager_replica.key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret_version" "rum_monitor_identity_pool_id" {
  secret_id     = aws_secretsmanager_secret.rum_monitor_identity_pool_id.id
  secret_string = aws_cognito_identity_pool.rum_monitor[0].id
  provider      = aws.eu_west_1
}
