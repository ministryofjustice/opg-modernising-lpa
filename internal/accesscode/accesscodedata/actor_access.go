package accesscodedata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"

type ActorAccess struct {
	PK           dynamo.ActorAccessKeyType
	SK           dynamo.MetadataKeyType
	ShareKey     dynamo.AccessKeyType
	ShareSortKey dynamo.AccessSortKeyType
}
