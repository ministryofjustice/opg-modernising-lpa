resource "aws_secretsmanager_secret" "private_jwt_key_base64" {
  name       = "private-jwt-key-base64"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id
  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "os_postcode_lookup_api_key" {
  name       = "os-postcode-lookup-api-key"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id
  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "cookie_session_keys" {
  name       = "cookie-session-keys"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id
  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "gov_uk_pay_api_key" {
  name       = "gov-uk-pay-api-key"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id
  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret" "gov_uk_notify_api_key" {
  name       = "gov-uk-notify-api-key"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id
  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

data "aws_secretsmanager_secret" "lpa_store_jwt_key" {
  name     = "opg-data-lpa-store/${data.aws_default_tags.global.tags.account-name}/jwt-key"
  provider = aws.management_eu_west_1
}

data "aws_secretsmanager_secret_version" "lpa_store_jwt_key" {
  secret_id = data.aws_secretsmanager_secret.lpa_store_jwt_key.id
  provider  = aws.management_eu_west_1
}

resource "aws_secretsmanager_secret" "lpa_store_jwt_secret_key" {
  name       = "lpa-store-jwt-secret-key"
  kms_key_id = module.secrets_manager_kms.eu_west_1_target_key_id

  replica {
    kms_key_id = module.secrets_manager_kms.eu_west_2_target_key_id
    region     = data.aws_region.eu_west_2.name
  }
  provider = aws.eu_west_1
}

resource "aws_secretsmanager_secret_version" "lpa_store_jwt_secret_key" {
  secret_id     = aws_secretsmanager_secret.lpa_store_jwt_secret_key.id
  secret_string = data.aws_secretsmanager_secret_version.lpa_store_jwt_key.secret_string
  provider      = aws.eu_west_1
}
