package inMemoryTokenBlacklist

import (
	"fmt"
	"geo/db/tokenBlacklist"
	"sync"
	"time"
)

type Blacklist struct {
	Skew time.Duration
	list map[string]time.Time
	mu   sync.RWMutex
}

func NewBlacklist(skew time.Duration) *Blacklist {
	return &Blacklist{
		list: make(map[string]time.Time, 100),
		Skew: skew,
	}
}

func (bl *Blacklist) Contains(jti string) bool {
	result := false
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if _, ok := bl.list[jti]; ok {
		result = true
	}
	bl.clean()
	return result
}

func (bl *Blacklist) Add(jti string, exp time.Time) error {
	if time.Now().UTC().After(exp.Add(bl.Skew)) {
		return fmt.Errorf("%w: %v", tokenBlacklist.Expired, jti)
	}
	bl.mu.Lock()
	defer bl.mu.Unlock()
	if _, ok := bl.list[jti]; ok {
		return fmt.Errorf("%w: %v", tokenBlacklist.JTIAlreadyExists, jti)
	}
	bl.list[jti] = exp
	bl.clean()
	return nil
}

func (bl *Blacklist) clean() {
	for token, exp := range bl.list {
		if time.Now().UTC().After(exp.Add(bl.Skew)) {
			delete(bl.list, token)
		}
	}
}
