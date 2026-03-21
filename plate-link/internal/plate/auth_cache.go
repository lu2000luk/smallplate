package plate

import (
	"fmt"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
)

type AuthDecision struct {
	PlateID string `json:"plate_id"`
	Key     string `json:"key"`
	Service string `json:"service"`
	Valid   bool   `json:"valid"`
}

type AuthCache struct {
	mu    sync.Mutex
	cache *lru.Cache[string, map[string]AuthDecision]
}

func NewAuthCache(size int) (*AuthCache, error) {
	cache, err := lru.New[string, map[string]AuthDecision](size)
	if err != nil {
		return nil, err
	}
	return &AuthCache{cache: cache}, nil
}

func (a *AuthCache) decisionKey(plateID string, service string) string {
	return fmt.Sprintf("%s|%s", plateID, service)
}

func (a *AuthCache) Get(plateID string, key string, service string) (AuthDecision, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	decisions, ok := a.cache.Get(key)
	if !ok {
		return AuthDecision{}, false
	}
	decision, ok := decisions[a.decisionKey(plateID, service)]
	return decision, ok
}

func (a *AuthCache) Add(decision AuthDecision) {
	a.mu.Lock()
	defer a.mu.Unlock()
	decisions, ok := a.cache.Get(decision.Key)
	if !ok {
		decisions = make(map[string]AuthDecision)
	}
	decisions[a.decisionKey(decision.PlateID, decision.Service)] = decision
	a.cache.Add(decision.Key, decisions)
}

func (a *AuthCache) Invalidate(rawKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache.Remove(rawKey)
}

func (a *AuthCache) InvalidatePlate(plateID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, rawKey := range a.cache.Keys() {
		decisions, ok := a.cache.Peek(rawKey)
		if !ok {
			continue
		}

		next := make(map[string]AuthDecision, len(decisions))
		for decisionKey, decision := range decisions {
			if decision.PlateID == plateID {
				continue
			}
			next[decisionKey] = decision
		}

		if len(next) == 0 {
			a.cache.Remove(rawKey)
			continue
		}

		a.cache.Add(rawKey, next)
	}
}
