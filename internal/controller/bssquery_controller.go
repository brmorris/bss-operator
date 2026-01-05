/*
Copyright 2026.

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
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
	bssclient "github.com/brmorris/bss-operator/internal/client"
)

const (
	// Condition types
	TypeAvailable = "Available"
	TypeDegraded  = "Degraded"

	// Condition reasons
	ReasonReconciling   = "Reconciling"
	ReasonQuerySuccess  = "QuerySuccess"
	ReasonQueryFailed   = "QueryFailed"
	ReasonInvalidConfig = "InvalidConfig"
)

// BSSQueryReconciler reconciles a BSSQuery object
type BSSQueryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=bss.localhost,resources=bssqueries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bss.localhost,resources=bssqueries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=bss.localhost,resources=bssqueries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BSSQueryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the BSSQuery instance
	bssQuery := &bssv1alpha1.BSSQuery{}
	if err := r.Get(ctx, req.NamespacedName, bssQuery); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("BSSQuery resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get BSSQuery")
		return ctrl.Result{}, err
	}

	// Set the status as Unknown when no status is available
	if len(bssQuery.Status.Conditions) == 0 {
		meta.SetStatusCondition(&bssQuery.Status.Conditions, metav1.Condition{
			Type:               TypeAvailable,
			Status:             metav1.ConditionUnknown,
			Reason:             ReasonReconciling,
			LastTransitionTime: metav1.Now(),
			Message:            "Starting reconciliation",
		})
		if err := r.Status().Update(ctx, bssQuery); err != nil {
			logger.Error(err, "Failed to update BSSQuery status")
			return ctrl.Result{}, err
		}

		// Re-fetch the BSSQuery after updating status
		if err := r.Get(ctx, req.NamespacedName, bssQuery); err != nil {
			logger.Error(err, "Failed to re-fetch BSSQuery")
			return ctrl.Result{}, err
		}
	}

	// Validate the query configuration
	if err := r.validateQuery(bssQuery); err != nil {
		meta.SetStatusCondition(&bssQuery.Status.Conditions, metav1.Condition{
			Type:               TypeDegraded,
			Status:             metav1.ConditionTrue,
			Reason:             ReasonInvalidConfig,
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
		})
		if err := r.Status().Update(ctx, bssQuery); err != nil {
			logger.Error(err, "Failed to update BSSQuery status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Execute the GraphQL query
	if err := r.executeQuery(ctx, bssQuery); err != nil {
		logger.Error(err, "Failed to execute query")
		meta.SetStatusCondition(&bssQuery.Status.Conditions, metav1.Condition{
			Type:               TypeDegraded,
			Status:             metav1.ConditionTrue,
			Reason:             ReasonQueryFailed,
			LastTransitionTime: metav1.Now(),
			Message:            fmt.Sprintf("Query failed: %v", err),
		})
		if err := r.Status().Update(ctx, bssQuery); err != nil {
			logger.Error(err, "Failed to update BSSQuery status")
			return ctrl.Result{}, err
		}
		// Requeue with a delay
		refreshInterval := time.Duration(bssQuery.Spec.RefreshInterval) * time.Second
		if refreshInterval == 0 {
			refreshInterval = 30 * time.Second
		}
		return ctrl.Result{RequeueAfter: refreshInterval}, nil
	}

	// Update status with success
	meta.SetStatusCondition(&bssQuery.Status.Conditions, metav1.Condition{
		Type:               TypeAvailable,
		Status:             metav1.ConditionTrue,
		Reason:             ReasonQuerySuccess,
		LastTransitionTime: metav1.Now(),
		Message:            "Query executed successfully",
	})

	meta.SetStatusCondition(&bssQuery.Status.Conditions, metav1.Condition{
		Type:               TypeDegraded,
		Status:             metav1.ConditionFalse,
		Reason:             ReasonQuerySuccess,
		LastTransitionTime: metav1.Now(),
		Message:            "Query executed successfully",
	})

	bssQuery.Status.ObservedGeneration = bssQuery.Generation
	now := metav1.Now()
	bssQuery.Status.LastQueryTime = &now

	if err := r.Status().Update(ctx, bssQuery); err != nil {
		logger.Error(err, "Failed to update BSSQuery status")
		return ctrl.Result{}, err
	}

	// Requeue after refresh interval
	refreshInterval := time.Duration(bssQuery.Spec.RefreshInterval) * time.Second
	if refreshInterval == 0 {
		refreshInterval = 30 * time.Second
	}

	logger.Info("Successfully reconciled BSSQuery", "requeueAfter", refreshInterval)
	return ctrl.Result{RequeueAfter: refreshInterval}, nil
}

// validateQuery validates the BSSQuery configuration
func (r *BSSQueryReconciler) validateQuery(bssQuery *bssv1alpha1.BSSQuery) error {
	if bssQuery.Spec.APIEndpoint == "" {
		return fmt.Errorf("APIEndpoint is required")
	}

	if bssQuery.Spec.Query == bssv1alpha1.QueryTypeCluster && bssQuery.Spec.ClusterID == "" {
		return fmt.Errorf("ClusterID is required for cluster query type")
	}

	return nil
}

// executeQuery executes the GraphQL query and updates the status
func (r *BSSQueryReconciler) executeQuery(ctx context.Context, bssQuery *bssv1alpha1.BSSQuery) error {
	logger := log.FromContext(ctx)

	// Create GraphQL client
	gqlClient := bssclient.NewGraphQLClient(bssQuery.Spec.APIEndpoint)

	switch bssQuery.Spec.Query {
	case bssv1alpha1.QueryTypeCluster:
		cluster, err := gqlClient.GetCluster(bssQuery.Spec.ClusterID)
		if err != nil {
			return fmt.Errorf("failed to get cluster: %w", err)
		}

		if cluster == nil {
			return fmt.Errorf("cluster not found: %s", bssQuery.Spec.ClusterID)
		}

		resultJSON, err := json.Marshal(cluster)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		bssQuery.Status.Result = string(resultJSON)
		bssQuery.Status.ClusterCount = 1
		logger.Info("Retrieved cluster", "id", cluster.ID, "name", cluster.Name, "state", cluster.State)

	case bssv1alpha1.QueryTypeClusters:
		clusters, err := gqlClient.ListClusters()
		if err != nil {
			return fmt.Errorf("failed to list clusters: %w", err)
		}

		resultJSON, err := json.Marshal(clusters)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		bssQuery.Status.Result = string(resultJSON)
		bssQuery.Status.ClusterCount = len(clusters)
		logger.Info("Retrieved clusters", "count", len(clusters))

	default:
		return fmt.Errorf("unknown query type: %s", bssQuery.Spec.Query)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BSSQueryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bssv1alpha1.BSSQuery{}).
		Complete(r)
}
