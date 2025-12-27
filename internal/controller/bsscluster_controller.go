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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
	"github.com/brmorris/bss-operator/internal/resources"
	"github.com/brmorris/bss-operator/internal/validation"
)

const (
	// Phase constants for BssCluster status
	phaseReconciling = "Reconciling"
	phaseFailed      = "Failed"
	phaseReady       = "Ready"
)

// BssClusterReconciler reconciles a BssCluster object
type BssClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Resource reconcilers
	statefulSetReconciler *resources.StatefulSetReconciler
	serviceReconciler     *resources.ServiceReconciler

	// Validator
	validator *validation.Validator
}

// NewBssClusterReconciler creates a new BssClusterReconciler with all dependencies
func NewBssClusterReconciler(c client.Client, scheme *runtime.Scheme) *BssClusterReconciler {
	return &BssClusterReconciler{
		Client:                c,
		Scheme:                scheme,
		statefulSetReconciler: resources.NewStatefulSetReconciler(c, scheme),
		serviceReconciler:     resources.NewServiceReconciler(c, scheme),
		validator:             validation.NewValidator(),
	}
}

// +kubebuilder:rbac:groups=bss.localhost,resources=bssclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bss.localhost,resources=bssclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=bss.localhost,resources=bssclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *BssClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the BssCluster instance
	var bssCluster bssv1alpha1.BssCluster
	if err := r.Get(ctx, req.NamespacedName, &bssCluster); err != nil {
		if errors.IsNotFound(err) {
			log.Info("BssCluster resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get BssCluster")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling BssCluster", "name", bssCluster.Name, "namespace", bssCluster.Namespace)

	// Validate the spec
	if err := r.validator.Validate(&bssCluster); err != nil {
		log.Error(err, "BssCluster validation failed")
		if statusErr := r.updateStatus(ctx, &bssCluster, phaseFailed); statusErr != nil {
			log.Error(statusErr, "Failed to update BssCluster status")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	// Update status to Reconciling
	if err := r.updateStatus(ctx, &bssCluster, phaseReconciling); err != nil {
		log.Error(err, "Failed to update BssCluster status to Reconciling")
		return ctrl.Result{}, err
	}

	// Reconcile all resources
	if err := r.reconcileResources(ctx, &bssCluster, log); err != nil {
		log.Error(err, "Failed to reconcile resources")
		_ = r.updateStatus(ctx, &bssCluster, phaseFailed)
		return ctrl.Result{}, err
	}

	// Update status to Ready
	if err := r.updateStatus(ctx, &bssCluster, phaseReady); err != nil {
		log.Error(err, "Failed to update BssCluster status to Ready")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled BssCluster", "name", bssCluster.Name)
	return ctrl.Result{}, nil
}

// reconcileResources reconciles all child resources
func (r *BssClusterReconciler) reconcileResources(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
	// Reconcile Service first (required for StatefulSet)
	if err := r.serviceReconciler.Reconcile(ctx, bssCluster, log); err != nil {
		return err
	}

	// Reconcile StatefulSet
	if err := r.statefulSetReconciler.Reconcile(ctx, bssCluster, log); err != nil {
		return err
	}

	// Add more resource reconcilers here as you expand:
	// - ConfigMaps
	// - Secrets
	// - PVCs
	// - Ingress
	// - etc.

	return nil
}

// updateStatus updates the status of the BssCluster
func (r *BssClusterReconciler) updateStatus(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, phase string) error {
	bssCluster.Status.Phase = phase
	return r.Status().Update(ctx, bssCluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BssClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bssv1alpha1.BssCluster{}).
		Named("bsscluster").
		Complete(r)
}
