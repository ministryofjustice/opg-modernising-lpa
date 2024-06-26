# Using Opensearch Dashboard dev mode

Some useful commands when working with the Opensearch dashboard in dev mode.

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
