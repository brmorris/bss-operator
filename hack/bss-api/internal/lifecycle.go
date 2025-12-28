package internal

import (
	"time"

	"github.com/brmorris/bss-operator/hack/bss-api/model"
)

func SimulateCreate(cluster *model.Cluster) {
	go func() {
		time.Sleep(20 * time.Second)

		cluster.State = model.StateReady
		cluster.ReadyReplicas = cluster.Replicas
		cluster.LastUpdateTime = time.Now()
	}()
}

func SimulateDelete(cluster *model.Cluster, onComplete func()) {
	go func() {
		time.Sleep(10 * time.Second)
		onComplete()
	}()
}
