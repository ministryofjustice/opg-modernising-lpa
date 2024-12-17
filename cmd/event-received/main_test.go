package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
)

var (
	expectedError = errors.New("err")
	ctx           = context.Background()

	testNow   = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn = func() time.Time { return testNow }

	testUuidString   = "a-uuid"
	testUuidStringFn = func() string { return testUuidString }
)

func TestFeeApprovedEventUnmarshalJSON(t *testing.T) {
	testcases := []pay.FeeType{
		pay.FullFee,
		pay.HalfFee,
		pay.QuarterFee,
		pay.NoFee,
	}

	for _, feeType := range testcases {
		t.Run(feeType.String(), func(t *testing.T) {
			event := feeApprovedEvent{
				UID:          "a",
				ApprovedType: feeType,
			}

			data, err := json.Marshal(event)
			assert.Nil(t, err)
			assert.Equal(t, fmt.Sprintf(`{"uid":"a","approvedType":"%s"}`, feeType.String()), string(data))

			var v feeApprovedEvent
			err = json.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, event, v)
		})
	}
}
