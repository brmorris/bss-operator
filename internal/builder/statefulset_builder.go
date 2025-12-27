/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package builder

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

// StatefulSetBuilder builds a StatefulSet for a BssCluster
type StatefulSetBuilder struct {
	bssCluster *bssv1alpha1.BssCluster
}

// NewStatefulSetBuilder creates a new StatefulSetBuilder
func NewStatefulSetBuilder(bssCluster *bssv1alpha1.BssCluster) *StatefulSetBuilder {
	return &StatefulSetBuilder{
		bssCluster: bssCluster,
	}
}

// Build constructs the StatefulSet
func (b *StatefulSetBuilder) Build() *appsv1.StatefulSet {
	replicas := b.getReplicas()
	labels := CommonLabels(b.bssCluster)
	selectorLabels := SelectorLabels(b.bssCluster)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.bssCluster.Name,
			Namespace: b.bssCluster.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &replicas,
			ServiceName: b.bssCluster.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: b.buildPodSpec(),
			},
			// VolumeClaimTemplates can be added here when you add PVC support
		},
	}
}

func (b *StatefulSetBuilder) buildPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "bss",
				Image: b.bssCluster.Spec.Image,
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 8080,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				// Resources, env vars, volume mounts can be added here
			},
		},
		// Volumes can be added here when you add ConfigMap/Secret support
	}
}

func (b *StatefulSetBuilder) getReplicas() int32 {
	if b.bssCluster.Spec.Replicas != nil {
		return *b.bssCluster.Spec.Replicas
	}
	return 1
}
