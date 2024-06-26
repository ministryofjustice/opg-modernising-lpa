resource "aws_iam_role" "opensearch_ingestion_service_role" {
  name               = "AWSServiceRoleForAmazonOpenSearchIngestionService"
  path               = "/aws-service-role/osis.amazonaws.com/"
  assume_role_policy = data.aws_iam_policy_document.opensearch_ingestion_service_role_assume_policy.json
  description        = "Allows Amazon OpenSearch Ingestion Service to access other AWS services on your behalf."
  provider           = aws.global
}

data "aws_iam_policy_document" "opensearch_ingestion_service_role_assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["osis.amazonaws.com"]
      type        = "Service"
    }
  }
  provider = aws.global
}

resource "aws_iam_role_policy_attachment" "opensearch_ingestion_service_role" {
  policy_arn = "arn:aws:iam::aws:policy/aws-service-role/AmazonOpenSearchIngestionServiceRolePolicy"
  role       = aws_iam_role.opensearch_ingestion_service_role.name
  provider   = aws.global
}
