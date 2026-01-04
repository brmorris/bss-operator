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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

var _ = Describe("BSSQuery Controller", func() {
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When reconciling a BSSQuery resource", func() {
		ctx := context.Background()

		It("should handle invalid configuration", func() {
			bssQuery := &bssv1alpha1.BSSQuery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-invalid-config",
					Namespace: "default",
				},
				Spec: bssv1alpha1.BSSQuerySpec{
					APIEndpoint: "", // Invalid: empty endpoint
					Query:       bssv1alpha1.QueryTypeCluster,
				},
			}

			Expect(k8sClient.Create(ctx, bssQuery)).Should(Succeed())

			// Wait for the status to be updated
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      bssQuery.Name,
					Namespace: bssQuery.Namespace,
				}, bssQuery)
				if err != nil {
					return false
				}
				return len(bssQuery.Status.Conditions) > 0
			}, timeout, interval).Should(BeTrue())

			// Clean up
			Expect(k8sClient.Delete(ctx, bssQuery)).Should(Succeed())
		})

		It("should handle missing ClusterID for cluster query", func() {
			bssQuery := &bssv1alpha1.BSSQuery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-missing-clusterid",
					Namespace: "default",
				},
				Spec: bssv1alpha1.BSSQuerySpec{
					APIEndpoint: "http://localhost:8880/graphql",
					Query:       bssv1alpha1.QueryTypeCluster,
					// Missing ClusterID
				},
			}

			Expect(k8sClient.Create(ctx, bssQuery)).Should(Succeed())

			// Wait for the status to be updated with error
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      bssQuery.Name,
					Namespace: bssQuery.Namespace,
				}, bssQuery)
				if err != nil {
					return false
				}
				for _, cond := range bssQuery.Status.Conditions {
					if cond.Type == TypeDegraded && cond.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			// Clean up
			Expect(k8sClient.Delete(ctx, bssQuery)).Should(Succeed())
		})

		It("should accept valid clusters query configuration", func() {
			bssQuery := &bssv1alpha1.BSSQuery{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-clusters-query",
					Namespace: "default",
				},
				Spec: bssv1alpha1.BSSQuerySpec{
					APIEndpoint:     "http://localhost:8880/graphql",
					Query:           bssv1alpha1.QueryTypeClusters,
					RefreshInterval: 30,
				},
			}

			Expect(k8sClient.Create(ctx, bssQuery)).Should(Succeed())

			// Wait for the resource to be created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      bssQuery.Name,
					Namespace: bssQuery.Namespace,
				}, bssQuery)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Verify the spec
			Expect(bssQuery.Spec.Query).To(Equal(bssv1alpha1.QueryTypeClusters))
			Expect(bssQuery.Spec.APIEndpoint).To(Equal("http://localhost:8880/graphql"))
			Expect(bssQuery.Spec.RefreshInterval).To(Equal(int32(30)))

			// Clean up
			Expect(k8sClient.Delete(ctx, bssQuery)).Should(Succeed())
		})
	})
})
