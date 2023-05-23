# UID service tester

During development or changes to the UID service it can be useful to quickly check if the service is working as intended. This script mimics the code found in /app to send a valid request to the UID service with required headers.

To run:

```bash
go build
aws-vault exec identity -- go run uid-tester
```

To vary the base URL for the service pass as an argument to the script:

```bash
aws-vault exec identity -- go run uid-tester -baseUrl=https://new.base.url
```
