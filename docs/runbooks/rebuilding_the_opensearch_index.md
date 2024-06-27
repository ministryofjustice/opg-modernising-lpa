# Rebuilding the Opensearch index

In the event that the Opensearch index needs to be rebuilt, the following steps should be followed.

These instructions are for the `test` environment with an index called `lpas_v2_test`. The same steps can be followed for the other environments, with the appropriate index name `lpas_v2_<environment name>`.

1. Delete the existing index

```shell
DELETE /lpas_v2_test
```

1. Create a new index with the correct mapping

```shell
PUT /lpas_v2_test
{
    "settings": {
        "index": {
            "number_of_shards": 1,
            "number_of_replicas": 1
        }
    },
    "mappings": {
        "properties": {
            "PK": {"type": "keyword"},
            "SK": {"type": "keyword"},
            "Donor.FirstNames": {"type": "keyword"},
            "Donor.LastName": {"type": "keyword"}
        }
    }
}
```

1. recreate the opensearch pipeline

We do this by using terraform to taint and recreate the pipeline.

In a shell, navigate to the `terraform/environment` directory and select the correct workspace:

```shell
tf workspace select <environment name>
```

(working with preproduction and production environments requires the breakglass role)

Mark the pipeline for recreation:

```shell
tf taint 'aws_osis_pipeline.lpas_stream[0]'
```

Then apply the changes:

```shell
tf apply
```

1. Reindexing

When the pipeline is created, it will trigger a dynamoDB export to S3. Once the export is finished, the pipeline will import the data into index. After the export processing is complete, the pipeline will switch to processing DynamoDB stream events if enabled.
