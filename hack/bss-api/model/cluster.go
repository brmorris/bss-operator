package model

import "time"

type ClusterState string

const (
	StateCreating ClusterState = "creating"
	StateReady    ClusterState = "ready"
	StateFailed   ClusterState = "failed"
	StateDeleting ClusterState = "deleting"
)

type Cluster struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Replicas       int32        `json:"replicas"`
	Version        string       `json:"version"`
	State          ClusterState `json:"state"`
	ReadyReplicas  int32        `json:"readyReplicas"`
	CreatedAt      time.Time    `json:"createdAt"`
	LastUpdateTime time.Time    `json:"lastUpdateTime"`
}
