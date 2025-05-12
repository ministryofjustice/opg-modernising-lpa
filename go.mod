module github.com/ministryofjustice/opg-modernising-lpa

go 1.24

toolchain go1.24.3

require (
	github.com/MicahParks/jwkset v0.9.5
	github.com/MicahParks/keyfunc/v3 v3.3.11
	github.com/aws/aws-lambda-go v1.48.0
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.19.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.44.3
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.43.1
	github.com/aws/aws-sdk-go-v2/service/eventbridge v1.39.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	github.com/aws/aws-sdk-go-v2/service/s3control v1.57.0
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.35.4
	github.com/aws/aws-sdk-go-v2/service/sqs v1.38.5
	github.com/aws/aws-sdk-go-v2/service/ssm v1.59.0
	github.com/aws/smithy-go v1.22.3
	github.com/felixge/httpsnoop v1.0.4
	github.com/gabriel-vasile/mimetype v1.4.9
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/sessions v1.4.0
	github.com/ministryofjustice/opg-go-common v1.91.0
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/opensearch-project/opensearch-go/v4 v4.4.0
	github.com/pact-foundation/pact-go/v2 v2.4.1
	github.com/stretchr/testify v1.10.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.35.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda v0.60.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig v0.60.0
	go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws v0.60.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0
	go.opentelemetry.io/contrib/propagators/aws v1.35.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	golang.org/x/mod v0.24.0
	golang.org/x/time v0.11.0
	golang.org/x/tools v0.33.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.34 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.34.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19 // indirect
	github.com/brunoscheufler/aws-ecs-metadata-go v0.0.0-20221221133751-67e37ae746cd // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/chigopher/pathlib v0.19.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/spf13/viper v1.20.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/vektra/mockery/v2 v2.53.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/detectors/aws/lambda v0.60.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/grpc v1.72.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	github.com/ministryofjustice/opg-modernising-lpa/cmd/enumerator
	github.com/pact-foundation/pact-go/v2
	github.com/vektra/mockery/v2
)
