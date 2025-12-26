package utils

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// APIKeyPool manages a pool of API keys with rotation and blacklisting
type APIKeyPool struct {
	keys         []string
	usageCounts  map[string]int
	lastUsedTime map[string]time.Time
	blacklist    map[string]time.Time
	mu           sync.RWMutex
}

// NewAPIKeyPool creates a new API key pool
func NewAPIKeyPool(keys []string) *APIKeyPool {
	if len(keys) == 0 {
		return nil
	}

	return &APIKeyPool{
		keys:         keys,
		usageCounts:  make(map[string]int),
		lastUsedTime: make(map[string]time.Time),
		blacklist:    make(map[string]time.Time),
	}
}

// GetRandomKey returns an available API key
// Implements smart selection: prefers less-used keys, avoids blacklisted keys
func (p *APIKeyPool) GetRandomKey() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Clean expired blacklist entries
	p.cleanBlacklist()

	// Filter available keys (not blacklisted)
	available := p.getAvailableKeys()
	if len(available) == 0 {
		return "", errors.New("no available API keys")
	}

	// Find minimum usage count
	minUsage := -1
	for _, key := range available {
		count := p.usageCounts[key]
		if minUsage == -1 || count < minUsage {
			minUsage = count
		}
	}

	// Select keys with minimum usage (top 50%)
	candidates := make([]string, 0)
	threshold := minUsage + (len(available) / 2)
	for _, key := range available {
		if p.usageCounts[key] <= threshold {
			candidates = append(candidates, key)
		}
	}

	// Random selection from candidates
	if len(candidates) == 0 {
		candidates = available
	}

	selectedKey := candidates[rand.Intn(len(candidates))]
	p.usageCounts[selectedKey]++
	p.lastUsedTime[selectedKey] = time.Now()

	return selectedKey, nil
}

// MarkSuccess marks a key as successfully used
func (p *APIKeyPool) MarkSuccess(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Key worked successfully - no action needed
	// Usage count already incremented in GetRandomKey
}

// MarkFailed marks a key as failed and temporarily blacklists it
func (p *APIKeyPool) MarkFailed(key string, retryAfter time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Add to blacklist with expiration time
	p.blacklist[key] = time.Now().Add(retryAfter)
}

// getAvailableKeys returns keys that are not blacklisted
// Must be called with lock held
func (p *APIKeyPool) getAvailableKeys() []string {
	available := make([]string, 0)
	now := time.Now()

	for _, key := range p.keys {
		if expireTime, exists := p.blacklist[key]; exists {
			if now.Before(expireTime) {
				// Still blacklisted
				continue
			}
		}
		available = append(available, key)
	}

	return available
}

// cleanBlacklist removes expired entries from blacklist
// Must be called with lock held
func (p *APIKeyPool) cleanBlacklist() {
	now := time.Now()
	for key, expireTime := range p.blacklist {
		if now.After(expireTime) {
			delete(p.blacklist, key)
		}
	}
}

// GetStats returns usage statistics
func (p *APIKeyPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	available := p.getAvailableKeys()

	return map[string]interface{}{
		"total_keys":     len(p.keys),
		"available_keys": len(available),
		"blacklisted":    len(p.keys) - len(available),
		"usage_counts":   p.usageCounts,
	}
}
