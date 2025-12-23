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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
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
	if err := r.validateBssCluster(&bssCluster); err != nil {
		log.Error(err, "BssCluster validation failed")
		bssCluster.Status.Phase = phaseFailed
		if statusErr := r.Status().Update(ctx, &bssCluster); statusErr != nil {
			log.Error(statusErr, "Failed to update BssCluster status")
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}

	// Update status to Reconciling
	bssCluster.Status.Phase = phaseReconciling
	if err := r.Status().Update(ctx, &bssCluster); err != nil {
		log.Error(err, "Failed to update BssCluster status to Reconciling")
		return ctrl.Result{}, err
	}

	// Reconcile the StatefulSet
	if err := r.reconcileStatefulSet(ctx, &bssCluster); err != nil {
		log.Error(err, "Failed to reconcile StatefulSet")
		bssCluster.Status.Phase = phaseFailed
		_ = r.Status().Update(ctx, &bssCluster)
		return ctrl.Result{}, err
	}

	// Reconcile the Service
	if err := r.reconcileService(ctx, &bssCluster); err != nil {
		log.Error(err, "Failed to reconcile Service")
		bssCluster.Status.Phase = phaseFailed
		_ = r.Status().Update(ctx, &bssCluster)
		return ctrl.Result{}, err
	}

	// Update status to Ready
	bssCluster.Status.Phase = phaseReady
	if err := r.Status().Update(ctx, &bssCluster); err != nil {
		log.Error(err, "Failed to update BssCluster status to Ready")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled BssCluster", "name", bssCluster.Name)
	return ctrl.Result{}, nil
}

// validateBssCluster validates the BssCluster spec
func (r *BssClusterReconciler) validateBssCluster(bssCluster *bssv1alpha1.BssCluster) error {
	if bssCluster.Spec.Image == "" {
		return fmt.Errorf("spec.image is required but not specified")
	}
	return nil
}

// reconcileStatefulSet creates or updates the StatefulSet for the BssCluster
func (r *BssClusterReconciler) reconcileStatefulSet(ctx context.Context, bssCluster *bssv1alpha1.BssCluster) error {
	log := logf.FromContext(ctx)

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bssCluster.Name,
			Namespace: bssCluster.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, statefulSet, func() error {
		// Set replicas (default to 1 if not specified)
		replicas := int32(1)
		if bssCluster.Spec.Replicas != nil {
			replicas = *bssCluster.Spec.Replicas
		}

		// Define the StatefulSet spec
		statefulSet.Spec = appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": bssCluster.Name,
				},
			},
			ServiceName: bssCluster.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": bssCluster.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bss",
							Image: bssCluster.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		}

		// Set the owner reference
		return controllerutil.SetControllerReference(bssCluster, statefulSet, r.Scheme)
	})

	if err != nil {
		return err
	}

	log.Info("StatefulSet reconciled", "operation", op, "name", statefulSet.Name)
	return nil
}

// reconcileService creates or updates the Service for the BssCluster
func (r *BssClusterReconciler) reconcileService(ctx context.Context, bssCluster *bssv1alpha1.BssCluster) error {
	log := logf.FromContext(ctx)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bssCluster.Name,
			Namespace: bssCluster.Namespace,
		},
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Define the Service spec
		service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{
				"app": bssCluster.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
			},
			ClusterIP: corev1.ClusterIPNone, // Headless service for StatefulSet
		}

		// Set the owner reference
		return controllerutil.SetControllerReference(bssCluster, service, r.Scheme)
	})

	if err != nil {
		return err
	}

	log.Info("Service reconciled", "operation", op, "name", service.Name)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BssClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bssv1alpha1.BssCluster{}).
		Named("bsscluster").
		Complete(r)
}
