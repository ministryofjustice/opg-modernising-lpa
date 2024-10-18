package main

import (
	"context"
	"errors"
	"time"
)

var (
	expectedError = errors.New("err")
	ctx           = context.Background()

	testNow   = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn = func() time.Time { return testNow }

	testUuidString   = "a-uuid"
	testUuidStringFn = func() string { return testUuidString }
)
