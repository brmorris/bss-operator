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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BSSQuerySpec defines the desired state of BSSQuery
type BSSQuerySpec struct {
	// APIEndpoint is the URL of the BSS API GraphQL endpoint
	// +kubebuilder:validation:Required
	APIEndpoint string `json:"apiEndpoint"`

	// Query specifies what to query from the BSS API
	// +kubebuilder:validation:Required
	Query BSSQueryType `json:"query"`

	// ClusterID is the cluster ID to query (for single cluster queries)
	// +optional
	ClusterID string `json:"clusterID,omitempty"`

	// RefreshInterval defines how often to refresh the query results (in seconds)
	// +kubebuilder:default=30
	// +optional
	RefreshInterval int32 `json:"refreshInterval,omitempty"`
}

// BSSQueryType defines the type of query to execute
// +kubebuilder:validation:Enum=cluster;clusters
type BSSQueryType string

const (
	QueryTypeCluster  BSSQueryType = "cluster"
	QueryTypeClusters BSSQueryType = "clusters"
)

// BSSQueryStatus defines the observed state of BSSQuery
type BSSQueryStatus struct {
	// LastQueryTime is the timestamp of the last successful query
	// +optional
	LastQueryTime *metav1.Time `json:"lastQueryTime,omitempty"`

	// Result contains the JSON result from the GraphQL query
	// +optional
	Result string `json:"result,omitempty"`

	// ClusterCount is the number of clusters returned (for list queries)
	// +optional
	ClusterCount int `json:"clusterCount,omitempty"`

	// Conditions represent the latest available observations of the BSSQuery's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed BSSQuery
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=bssq
//+kubebuilder:printcolumn:name="Query Type",type=string,JSONPath=`.spec.query`
//+kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.spec.apiEndpoint`
//+kubebuilder:printcolumn:name="Last Query",type=date,JSONPath=`.status.lastQueryTime`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BSSQuery is the Schema for the bssqueries API
type BSSQuery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BSSQuerySpec   `json:"spec,omitempty"`
	Status BSSQueryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BSSQueryList contains a list of BSSQuery
type BSSQueryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BSSQuery `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BSSQuery{}, &BSSQueryList{})
}
