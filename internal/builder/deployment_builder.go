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

// DeploymentBuilder builds a Deployment for a BssCluster
type DeploymentBuilder struct {
	bssCluster *bssv1alpha1.BssCluster
}

// NewDeploymentBuilder creates a new DeploymentBuilder
func NewDeploymentBuilder(bssCluster *bssv1alpha1.BssCluster) *DeploymentBuilder {
	return &DeploymentBuilder{
		bssCluster: bssCluster,
	}
}

// Build constructs the Deployment for bss-api
func (b *DeploymentBuilder) Build() *appsv1.Deployment {
	replicas := b.getReplicas()
	labels := CommonLabels(b.bssCluster)
	selectorLabels := SelectorLabels(b.bssCluster)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.bssCluster.Name,
			Namespace: b.bssCluster.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: b.buildPodSpec(),
			},
		},
	}
}

func (b *DeploymentBuilder) buildPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "bss-api",
				Image: b.getImage(),
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 8880,
						Protocol:      corev1.ProtocolTCP,
					},
				},
			},
		},
	}
}

func (b *DeploymentBuilder) getReplicas() int32 {
	if b.bssCluster.Spec.Replicas != nil {
		return *b.bssCluster.Spec.Replicas
	}
	return 1
}

func (b *DeploymentBuilder) getImage() string {
	return "bss-api:" + b.bssCluster.Spec.Version
}
