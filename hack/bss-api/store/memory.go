package store

import (
	"sync"

	"github.com/brmorris/bss-operator/hack/bss-api/model"
)

type MemoryStore struct {
	mu       sync.RWMutex
	clusters map[string]*model.Cluster
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		clusters: make(map[string]*model.Cluster),
	}
}

func (s *MemoryStore) Create(cluster *model.Cluster) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clusters[cluster.ID] = cluster
}

func (s *MemoryStore) Get(id string) (*model.Cluster, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.clusters[id]
	return c, ok
}

func (s *MemoryStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clusters, id)
}
