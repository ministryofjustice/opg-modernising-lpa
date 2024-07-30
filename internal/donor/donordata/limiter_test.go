package donordata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(time.Minute, 5, 10)

	assert.Equal(t, time.Minute, limiter.TokenPer)
	assert.Equal(t, float64(5), limiter.Tokens)
	assert.Equal(t, float64(10), limiter.MaxTokens)
	assert.WithinDuration(t, time.Now(), limiter.TokensAt, time.Millisecond)
}

func TestLimiter(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		limiter *Limiter
		allowed bool
	}{
		"has a token": {
			limiter: &Limiter{
				TokenPer:  time.Minute,
				Tokens:    1,
				MaxTokens: 1,
				TokensAt:  time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
			allowed: true,
		},
		"has no tokens": {
			limiter: &Limiter{
				TokenPer:  time.Minute,
				Tokens:    0,
				MaxTokens: 1,
				TokensAt:  now,
			},
			allowed: false,
		},
		"gets a token on refresh": {
			limiter: &Limiter{
				TokenPer:  time.Minute,
				Tokens:    0,
				MaxTokens: 1,
				TokensAt:  now.Add(-time.Minute),
			},
			allowed: true,
		},
		"gets a partial token on refresh": {
			limiter: &Limiter{
				TokenPer:  time.Minute,
				Tokens:    0.5,
				MaxTokens: 1,
				TokensAt:  now.Add(-time.Second * 30),
			},
			allowed: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.allowed, tc.limiter.Allow(now))
		})
	}
}

func TestLimiterBurst(t *testing.T) {
	now := time.Now()
	limiter := &Limiter{TokenPer: time.Second, Tokens: 3, MaxTokens: 5, TokensAt: now.Add(-time.Second)}

	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.False(t, limiter.Allow(now))
}

func TestLimiterMax(t *testing.T) {
	now := time.Now()
	limiter := &Limiter{TokenPer: time.Second, Tokens: 0, MaxTokens: 5, TokensAt: now.Add(-10 * time.Second)}

	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.True(t, limiter.Allow(now))
	assert.False(t, limiter.Allow(now))
}
