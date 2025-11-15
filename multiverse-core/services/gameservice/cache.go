package gameservice

import (
	"sync"
	"time"
	"multiverse-core/internal/entity"
)

// EntityCacheKey представляет ключ для кэширования сущностей
type EntityCacheKey struct {
	EntityID string
	WorldID  string
}

// EntityCache реализует кэш сущностей с TTL
type EntityCache struct {
	cache map[EntityCacheKey]*CachedEntity
	mutex sync.RWMutex
	ttl   time.Duration
}

// CachedEntity представляет сущность в кэше с метаданными
type CachedEntity struct {
	Entity    *entity.Entity
	Timestamp time.Time
}

// NewEntityCache создает новый кэш сущностей с заданным TTL
func NewEntityCache(ttl time.Duration) *EntityCache {
	cache := &EntityCache{
		cache: make(map[EntityCacheKey]*CachedEntity),
		ttl:   ttl,
	}
	
	// Запускаем горутину для очистки устаревших записей
	go cache.cleanup()
	
	return cache
}

// Get возвращает сущность из кэша, если она существует и не устарела
func (ec *EntityCache) Get(entityID, worldID string) (*entity.Entity, bool) {
	key := EntityCacheKey{EntityID: entityID, WorldID: worldID}
	
	ec.mutex.RLock()
	defer ec.mutex.RUnlock()
	
	if cached, exists := ec.cache[key]; exists {
		if time.Since(cached.Timestamp) < ec.ttl {
			return cached.Entity, true
		}
	}
	
	return nil, false
}

// Set добавляет сущность в кэш
func (ec *EntityCache) Set(entityID, worldID string, entity *entity.Entity) {
	key := EntityCacheKey{EntityID: entityID, WorldID: worldID}
	
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	ec.cache[key] = &CachedEntity{
		Entity:    entity,
		Timestamp: time.Now(),
	}
}

// Delete удаляет сущность из кэша
func (ec *EntityCache) Delete(entityID, worldID string) {
	key := EntityCacheKey{EntityID: entityID, WorldID: worldID}
	
	ec.mutex.Lock()
	defer ec.mutex.Unlock()
	
	delete(ec.cache, key)
}

// cleanup периодически удаляет устаревшие записи из кэша
func (ec *EntityCache) cleanup() {
	ticker := time.NewTicker(ec.ttl)
	defer ticker.Stop()
	
	for range ticker.C {
		ec.mutex.Lock()
		now := time.Now()
		for key, cached := range ec.cache {
			if now.Sub(cached.Timestamp) >= ec.ttl {
				delete(ec.cache, key)
			}
		}
		ec.mutex.Unlock()
	}
}