package cache

import (
	"container/list"
	"context"
	"hash/fnv"
	"sync"
	"time"
)

type item struct {
	entry   Entry
	size    int64
	element *list.Element
}
type shard struct {
	mu    sync.RWMutex
	items map[string]*item
	lru   *list.List
	bytes int64
}

type MemoryStore struct {
	opt    Options
	shards []shard
	stop   chan struct{}
	once   sync.Once
}

func NewMemoryStore(opt Options) *MemoryStore {
	opt = opt.normalize()
	m := &MemoryStore{opt: opt, shards: make([]shard, opt.Shards), stop: make(chan struct{})}
	for i := range m.shards {
		m.shards[i] = shard{items: map[string]*item{}, lru: list.New()}
	}
	go m.janitor()
	return m
}
func (m *MemoryStore) shardFor(key string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return &m.shards[int(h.Sum32())%len(m.shards)]
}
func (m *MemoryStore) Get(ctx context.Context, key string) (Entry, error) {
	st := time.Now()
	s := m.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[key]
	if !ok {
		m.opt.Metrics.IncMiss("memory")
		return Entry{}, ErrNotFound
	}
	if it.entry.Expired(time.Now()) {
		delete(s.items, key)
		s.lru.Remove(it.element)
		s.bytes -= it.size
		m.opt.Metrics.IncMiss("memory")
		return Entry{}, ErrNotFound
	}
	s.lru.MoveToFront(it.element)
	m.opt.Metrics.IncHit("memory")
	m.opt.Metrics.ObserveLatency("memory", "get", time.Since(st))
	return cloneEntry(it.entry), nil
}
func (m *MemoryStore) Set(ctx context.Context, e Entry, ttl time.Duration) error {
	st := time.Now()
	if ttl < 0 {
		return ErrInvalidTTL
	}
	if ttl == 0 {
		ttl = m.opt.DefaultTTL
	}
	if ttl > 0 {
		e.ExpiresAt = time.Now().Add(ttl)
	}
	e = cloneEntry(e)
	s := m.shardFor(e.Key)
	s.mu.Lock()
	defer s.mu.Unlock()
	size := int64(len(e.Key) + len(e.Value))
	if old := s.items[e.Key]; old != nil {
		old.entry = e
		s.bytes += size - old.size
		old.size = size
		s.lru.MoveToFront(old.element)
	} else {
		el := s.lru.PushFront(e.Key)
		s.items[e.Key] = &item{entry: e, size: size, element: el}
		s.bytes += size
	}
	m.evictLocked(s)
	m.opt.Metrics.ObserveLatency("memory", "set", time.Since(st))
	return nil
}
func (m *MemoryStore) Delete(ctx context.Context, key string) error {
	s := m.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	if it := s.items[key]; it != nil {
		delete(s.items, key)
		s.lru.Remove(it.element)
		s.bytes -= it.size
	}
	return nil
}
func (m *MemoryStore) Has(ctx context.Context, key string) (bool, error) {
	_, err := m.Get(ctx, key)
	if err == ErrNotFound {
		return false, nil
	}
	return err == nil, err
}
func (m *MemoryStore) Close() error { m.once.Do(func() { close(m.stop) }); return nil }
func (m *MemoryStore) evictLocked(s *shard) {
	maxE := m.opt.MaxEntries / m.opt.Shards
	if maxE < 1 {
		maxE = 1
	}
	maxB := m.opt.MaxBytes / int64(m.opt.Shards)
	for len(s.items) > maxE || s.bytes > maxB {
		back := s.lru.Back()
		if back == nil {
			return
		}
		k := back.Value.(string)
		it := s.items[k]
		delete(s.items, k)
		s.lru.Remove(back)
		s.bytes -= it.size
	}
}
func (m *MemoryStore) janitor() {
	t := time.NewTicker(m.opt.CleanupInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			m.cleanup()
		case <-m.stop:
			return
		}
	}
}
func (m *MemoryStore) cleanup() {
	now := time.Now()
	for i := range m.shards {
		s := &m.shards[i]
		s.mu.Lock()
		for k, it := range s.items {
			if it.entry.Expired(now) {
				delete(s.items, k)
				s.lru.Remove(it.element)
				s.bytes -= it.size
			}
		}
		s.mu.Unlock()
	}
}
func cloneEntry(e Entry) Entry {
	e.Value = append([]byte(nil), e.Value...)
	if e.Metadata != nil {
		m := make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			m[k] = v
		}
		e.Metadata = m
	}
	return e
}
