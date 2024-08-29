package dynamo

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/stretchr/testify/assert"
)

func TestPKType(t *testing.T) {
	pk := LpaKey("hello")
	pkType := WrapPK(pk)

	assert.Equal(t, pk, pkType.Unwrap())

	t.Run("json", func(t *testing.T) {
		marshalledPK, _ := json.Marshal(pk)
		marshalledPKType, _ := json.Marshal(pkType)

		assert.Equal(t, marshalledPK, marshalledPKType)

		var unmarshalled PKType
		_ = json.Unmarshal(marshalledPKType, &unmarshalled)
		assert.Equal(t, pk, unmarshalled.Unwrap())
	})

	t.Run("attributevalue", func(t *testing.T) {
		marshalledPK, _ := attributevalue.Marshal(pk)
		marshalledPKType, _ := attributevalue.Marshal(pkType)

		assert.Equal(t, marshalledPK, marshalledPKType)

		var unmarshalled PKType
		_ = attributevalue.Unmarshal(marshalledPKType, &unmarshalled)
		assert.Equal(t, pk, unmarshalled.Unwrap())
	})
}

func TestSKType(t *testing.T) {
	sk := DonorKey("hello")
	skType := WrapSK(sk)

	assert.Equal(t, sk, skType.Unwrap())

	t.Run("json", func(t *testing.T) {
		marshalledSK, _ := json.Marshal(sk)
		marshalledSKType, _ := json.Marshal(skType)

		assert.Equal(t, marshalledSK, marshalledSKType)

		var unmarshalled SKType
		_ = json.Unmarshal(marshalledSKType, &unmarshalled)
		assert.Equal(t, sk, unmarshalled.Unwrap())
	})

	t.Run("attributevalue", func(t *testing.T) {
		marshalledSK, _ := attributevalue.Marshal(sk)
		marshalledSKType, _ := attributevalue.Marshal(skType)

		assert.Equal(t, marshalledSK, marshalledSKType)

		var unmarshalled SKType
		_ = attributevalue.Unmarshal(marshalledSKType, &unmarshalled)
		assert.Equal(t, sk, unmarshalled.Unwrap())
	})
}
