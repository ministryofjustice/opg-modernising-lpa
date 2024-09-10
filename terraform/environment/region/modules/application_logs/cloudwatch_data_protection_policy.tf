resource "aws_cloudwatch_log_data_protection_policy" "application_logs" {
  log_group_name = aws_cloudwatch_log_group.application_logs.name
  policy_document = jsonencode({
    Name    = "data-protection-${data.aws_default_tags.current.tags.environment-name}-application-logs"
    Version = "2021-06-01"

    "Statement" : [
      {
        "Sid" : "audit-policy",
        "DataIdentifier" : [
          "arn:aws:dataprotection::aws:data-identifier/EmailAddress"
        ],
        "Operation" : {
          "Audit" : {
            "FindingsDestination" : {}
          }
        }
      },
      {
        "Sid" : "redact-policy",
        "DataIdentifier" : [
          "arn:aws:dataprotection::aws:data-identifier/EmailAddress"
        ],
        "Operation" : {
          "Deidentify" : {
            "MaskConfig" : {}
          }
        }
      }
    ]
  })
  provider = aws.region
}
