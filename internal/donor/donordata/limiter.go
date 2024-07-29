package donordata

import (
	"sync"
	"time"
)

// Limiter is a basic rate limiter that can be serialised.
type Limiter struct {
	TokenPer  time.Duration
	MaxTokens float64

	mu       sync.Mutex
	Tokens   float64
	TokensAt time.Time
}

func NewLimiter(tokenPer time.Duration, initialTokens, maxTokens float64) *Limiter {
	return &Limiter{
		TokenPer:  tokenPer,
		MaxTokens: maxTokens,
		Tokens:    initialTokens,
		TokensAt:  time.Now(),
	}
}

func (l *Limiter) Allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	elapsed := now.Sub(l.TokensAt)
	l.Tokens += elapsed.Seconds() / l.TokenPer.Seconds()
	l.TokensAt = now

	if l.Tokens > l.MaxTokens {
		l.Tokens = l.MaxTokens
	}

	if l.Tokens >= 1 {
		l.Tokens--
		return true
	}

	return false
}
