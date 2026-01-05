package graphql

import (
	"time"

	"github.com/brmorris/bss-operator/hack/bss-api/internal"
	"github.com/brmorris/bss-operator/hack/bss-api/model"
	"github.com/brmorris/bss-operator/hack/bss-api/store"
	"github.com/google/uuid"
)

func createCluster(s *store.MemoryStore, name string, replicas int32, version string) *model.Cluster {
	cluster := &model.Cluster{
		ID:             uuid.NewString(),
		Name:           name,
		Replicas:       replicas,
		Version:        version,
		State:          model.StateCreating,
		CreatedAt:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	s.Create(cluster)
	internal.SimulateCreate(cluster)

	return cluster
}

func deleteCluster(s *store.MemoryStore, id string) bool {
	cluster, ok := s.Get(id)
	if !ok {
		return false
	}

	cluster.State = model.StateDeleting

	internal.SimulateDelete(cluster, func() {
		s.Delete(id)
	})

	return true
}
