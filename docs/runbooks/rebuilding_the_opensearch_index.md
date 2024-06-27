# Rebuilding the Opensearch index

In the event that the Opensearch index needs to be rebuilt, the following steps should be followed.

1. Delete the existing index

```shell
# delete an index
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

In a shell, navigate to the `terraform/environment` directory and run the following command:

```shell
tf taint 'aws_osis_pipeline.lpas_stream[0]'
```

Then apply the changes:

```shell
tf apply
```

```shell
# list indices
GET _cat/indices?v

# create an index with the correct mapping
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

# show the mapping for an index
GET lpas_v2_test/_mapping

# delete an index
DELETE /lpas_v2_test

# return documents from an index
GET lpas_v2_test/_search
{"query":{"match_all":{}}}
```
