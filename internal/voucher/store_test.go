package voucher

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
)

var ctx = context.WithValue(context.Background(), "a", "b")

func TestNewStore(t *testing.T) {
	s := NewStore("a")

	assert.Equal(t, "a", s.dynamoClient)
}

func TestStoreCreate(t *testing.T) {
	store := Store{}
	result, err := store.Create(ctx, sharecodedata.Link{}, "")
	assert.Nil(t, err)
	assert.Nil(t, result)
}
