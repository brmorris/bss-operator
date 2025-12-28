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

package resources

import (
	"context"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
	"github.com/brmorris/bss-operator/internal/builder"
)

// DeploymentReconciler handles Deployment reconciliation
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewDeploymentReconciler creates a new DeploymentReconciler
func NewDeploymentReconciler(c client.Client, scheme *runtime.Scheme) *DeploymentReconciler {
	return &DeploymentReconciler{
		Client: c,
		Scheme: scheme,
	}
}

// Reconcile ensures the Deployment exists and matches the desired state
func (r *DeploymentReconciler) Reconcile(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
	desired := builder.NewDeploymentBuilder(bssCluster).Build()

	// Try to get the existing Deployment
	existing := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      desired.Name,
		Namespace: desired.Namespace,
	}, existing)

	if err != nil {
		if errors.IsNotFound(err) {
			return r.create(ctx, bssCluster, desired, log)
		}
		return err
	}

	return r.update(ctx, bssCluster, existing, desired, log)
}

func (r *DeploymentReconciler) create(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, deployment *appsv1.Deployment, log logr.Logger) error {
	// Set owner reference
	if err := controllerutil.SetControllerReference(bssCluster, deployment, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating Deployment", "name", deployment.Name)
	if err := r.Create(ctx, deployment); err != nil {
		return err
	}

	log.Info("Deployment created successfully", "name", deployment.Name)
	return nil
}

func (r *DeploymentReconciler) update(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, existing, desired *appsv1.Deployment, log logr.Logger) error {
	// Copy resource version and other metadata that should be preserved
	desired.ResourceVersion = existing.ResourceVersion

	// Set owner reference
	if err := controllerutil.SetControllerReference(bssCluster, desired, r.Scheme); err != nil {
		return err
	}

	// Check if update is needed (simple comparison for now)
	if r.needsUpdate(existing, desired) {
		log.Info("Updating Deployment", "name", desired.Name)
		if err := r.Update(ctx, desired); err != nil {
			return err
		}
		log.Info("Deployment updated successfully", "name", desired.Name)
	} else {
		log.V(1).Info("Deployment already up to date", "name", desired.Name)
	}

	return nil
}

func (r *DeploymentReconciler) needsUpdate(existing, desired *appsv1.Deployment) bool {
	// Compare replicas
	if existing.Spec.Replicas == nil || desired.Spec.Replicas == nil {
		return true
	}
	if *existing.Spec.Replicas != *desired.Spec.Replicas {
		return true
	}

	// Compare image
	if len(existing.Spec.Template.Spec.Containers) == 0 || len(desired.Spec.Template.Spec.Containers) == 0 {
		return true
	}
	if existing.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image {
		return true
	}

	return false
}
