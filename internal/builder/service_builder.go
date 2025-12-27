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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

// ServiceBuilder builds a Service for a BssCluster
type ServiceBuilder struct {
	bssCluster *bssv1alpha1.BssCluster
}

// NewServiceBuilder creates a new ServiceBuilder
func NewServiceBuilder(bssCluster *bssv1alpha1.BssCluster) *ServiceBuilder {
	return &ServiceBuilder{
		bssCluster: bssCluster,
	}
}

// Build constructs the Service
func (b *ServiceBuilder) Build() *corev1.Service {
	labels := CommonLabels(b.bssCluster)
	selectorLabels := SelectorLabels(b.bssCluster)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.bssCluster.Name,
			Namespace: b.bssCluster.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: corev1.ClusterIPNone, // Headless service for StatefulSet
			Selector:  selectorLabels,
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}
