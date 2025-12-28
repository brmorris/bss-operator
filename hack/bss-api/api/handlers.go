package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brmorris/bss-operator/hack/bss-api/internal"
	"github.com/brmorris/bss-operator/hack/bss-api/model"
	"github.com/brmorris/bss-operator/hack/bss-api/store"
	"github.com/google/uuid"
)

type Server struct {
	store *store.MemoryStore
}

func NewServer(store *store.MemoryStore) *Server {
	return &Server{store: store}
}

func (s *Server) CreateCluster(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Replicas int32  `json:"replicas"`
		Version  string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cluster := &model.Cluster{
		ID:             uuid.NewString(),
		Name:           req.Name,
		Replicas:       req.Replicas,
		Version:        req.Version,
		State:          model.StateCreating,
		CreatedAt:      time.Now(),
		LastUpdateTime: time.Now(),
	}

	s.store.Create(cluster)
	internal.SimulateCreate(cluster)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cluster)
}

func (s *Server) GetCluster(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cluster, ok := s.store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	_ = json.NewEncoder(w).Encode(cluster)
}

func (s *Server) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cluster, ok := s.store.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}

	cluster.State = model.StateDeleting

	internal.SimulateDelete(cluster, func() {
		s.store.Delete(id)
	})

	w.WriteHeader(http.StatusAccepted)
}
