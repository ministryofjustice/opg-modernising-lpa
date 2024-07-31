# Working with AWS Inspector SBOM exports

## request SBOM export

```bash
aws-vault exec management-operator -- \
    aws inspector2 create-sbom-export \
    --report-format SPDX_2_3 \
    --resource-filter-criteria file://filter_criteria.json \
    --s3-destination bucketName=opg-aws-inspector-sbom,keyPrefix=v0.1323.0,kmsKeyArn=arn:aws:kms:eu-west-1:311462405659:key/mrk-1899eeb57e6045d1a85310e1edda47c9
```

## get status of export

```bash
aws-vault exec management-operator -- \
    aws inspector2 get-sbom-export \
--report-id ba783153-5dc6-40ae-a9c9-9b48b232ec7b
```

## cancel status of export

```bash
aws-vault exec management-operator -- \
    aws inspector2 cancel-sbom-export \
--report-id 516b3fd1-881a-41a8-9592-d0fa70207e0f
```

## download the export

```bash
aws-vault exec management-operator -- \
    aws s3 cp s3://opg-aws-inspector-sbom/latest/SPDX_2_3_outputs_6ebd4d72-7eca-4693-bfbe-fb078ac11a6e/account=311462405659/resource=AWS_ECR_CONTAINER_IMAGE/ . --recursive
```
