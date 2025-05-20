package reuse

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func OrStringUnmarshalDynamoDBAttributeValue(t *testing.T) {
	var i orString[int]
	assert.Nil(t, attributevalue.Unmarshal(&types.AttributeValueMemberN{Value: "5"}, &i))
	assert.Equal(t, 5, i.v)
	assert.Equal(t, "", i.s)

	var s orString[int]
	assert.Nil(t, attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "hey"}, &s))
	assert.Equal(t, 0, i.v)
	assert.Equal(t, "hey", i.s)

	err := attributevalue.Unmarshal(&types.AttributeValueMemberBOOL{Value: true}, &s)
	assert.Error(t, err)
}
