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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
	"github.com/brmorris/bss-operator/internal/builder"
)

// ServiceReconciler handles Service reconciliation
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewServiceReconciler creates a new ServiceReconciler
func NewServiceReconciler(c client.Client, scheme *runtime.Scheme) *ServiceReconciler {
	return &ServiceReconciler{
		Client: c,
		Scheme: scheme,
	}
}

// Reconcile ensures the Service exists and matches the desired state
func (r *ServiceReconciler) Reconcile(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
	desired := builder.NewServiceBuilder(bssCluster).Build()

	// Try to get the existing Service
	existing := &corev1.Service{}
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

func (r *ServiceReconciler) create(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, service *corev1.Service, log logr.Logger) error {
	// Set owner reference
	if err := controllerutil.SetControllerReference(bssCluster, service, r.Scheme); err != nil {
		return err
	}

	log.Info("Creating Service", "name", service.Name)
	if err := r.Create(ctx, service); err != nil {
		return err
	}

	log.Info("Service created successfully", "name", service.Name)
	return nil
}

func (r *ServiceReconciler) update(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, existing, desired *corev1.Service, log logr.Logger) error {
	// Preserve immutable fields
	desired.Spec.ClusterIP = existing.Spec.ClusterIP
	desired.ResourceVersion = existing.ResourceVersion

	// Set owner reference
	if err := controllerutil.SetControllerReference(bssCluster, desired, r.Scheme); err != nil {
		return err
	}

	// Check if update is needed
	if r.needsUpdate(existing, desired) {
		log.Info("Updating Service", "name", desired.Name)
		if err := r.Update(ctx, desired); err != nil {
			return err
		}
		log.Info("Service updated successfully", "name", desired.Name)
	} else {
		log.V(1).Info("Service is up to date", "name", desired.Name)
	}

	return nil
}

// needsUpdate determines if the Service needs to be updated
func (r *ServiceReconciler) needsUpdate(existing, desired *corev1.Service) bool {
	// Compare ports
	if len(existing.Spec.Ports) != len(desired.Spec.Ports) {
		return true
	}

	for i := range existing.Spec.Ports {
		if existing.Spec.Ports[i].Port != desired.Spec.Ports[i].Port {
			return true
		}
	}

	// Add more sophisticated comparison as needed
	return false
}

// Delete removes the Service if it exists
func (r *ServiceReconciler) Delete(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      bssCluster.Name,
		Namespace: bssCluster.Namespace,
	}, service)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil // Already deleted
		}
		return err
	}

	log.Info("Deleting Service", "name", service.Name)
	return r.Client.Delete(ctx, service)
}
