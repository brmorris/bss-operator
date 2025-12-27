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

package validation

import (
	"fmt"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

// Validator validates BssCluster resources
type Validator struct{}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs validation on a BssCluster
func (v *Validator) Validate(bssCluster *bssv1alpha1.BssCluster) error {
	if err := v.validateSpec(bssCluster); err != nil {
		return err
	}

	// Add more validation as needed
	return nil
}

func (v *Validator) validateSpec(bssCluster *bssv1alpha1.BssCluster) error {
	// Validate image
	if bssCluster.Spec.Image == "" {
		return fmt.Errorf("spec.image is required but not specified")
	}

	// Validate replicas
	if bssCluster.Spec.Replicas != nil && *bssCluster.Spec.Replicas < 0 {
		return fmt.Errorf("spec.replicas must be non-negative")
	}

	// Add more spec validation as you add fields
	return nil
}
