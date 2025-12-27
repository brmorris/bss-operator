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
	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

const (
	// Label keys
	LabelApp       = "app.kubernetes.io/name"
	LabelInstance  = "app.kubernetes.io/instance"
	LabelVersion   = "app.kubernetes.io/version"
	LabelComponent = "app.kubernetes.io/component"
	LabelPartOf    = "app.kubernetes.io/part-of"
	LabelManagedBy = "app.kubernetes.io/managed-by"
)

// CommonLabels generates the standard set of labels for all resources
func CommonLabels(bssCluster *bssv1alpha1.BssCluster) map[string]string {
	return map[string]string{
		LabelApp:       "bss-cluster",
		LabelInstance:  bssCluster.Name,
		LabelPartOf:    "bss-operator",
		LabelManagedBy: "bss-operator",
	}
}

// SelectorLabels generates labels used for selectors (subset of common labels)
func SelectorLabels(bssCluster *bssv1alpha1.BssCluster) map[string]string {
	return map[string]string{
		LabelApp:      "bss-cluster",
		LabelInstance: bssCluster.Name,
	}
}

// MergeLabels merges multiple label maps with later maps taking precedence
func MergeLabels(labelMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, labels := range labelMaps {
		for k, v := range labels {
			result[k] = v
		}
	}
	return result
}
