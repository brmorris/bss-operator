# Enterprise Controller Architecture

## Overview

This operator follows an enterprise-grade, production-ready controller pattern that separates concerns and makes it easy to add new Kubernetes resources.

## Architecture

```
internal/
├── controller/              # Main reconciliation logic
│   ├── bsscluster_controller.go
│   └── bsscluster_controller_test.go
├── resources/              # Resource-specific reconcilers
│   ├── statefulset.go
│   ├── service.go
│   ├── configmap.go       # Add your own
│   └── ...
├── builder/               # Pure functions for building K8s objects
│   ├── statefulset_builder.go
│   ├── service_builder.go
│   ├── labels.go
│   └── ...
└── validation/            # Validation logic
    └── validator.go
```

## Key Design Principles

### 1. Separation of Concerns
- **Controller**: Orchestrates the reconciliation loop
- **Resources**: Handle CRUD operations for specific K8s resources
- **Builders**: Construct desired state (pure functions)
- **Validation**: Validates CR specifications

### 2. Single Responsibility
Each reconciler handles exactly one Kubernetes resource type.

### 3. Testability
- Builders are pure functions (easy to unit test)
- Resource reconcilers are independent
- Controller orchestration is testable

### 4. Scalability
Adding new resources is straightforward - just add new reconciler and builder.

## Adding a New Resource

### Example: Adding ConfigMap Support

#### Step 1: Create the Builder

Create `internal/builder/configmap_builder.go`:

```go
package builder

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

type ConfigMapBuilder struct {
    bssCluster *bssv1alpha1.BssCluster
}

func NewConfigMapBuilder(bssCluster *bssv1alpha1.BssCluster) *ConfigMapBuilder {
    return &ConfigMapBuilder{bssCluster: bssCluster}
}

func (b *ConfigMapBuilder) Build() *corev1.ConfigMap {
    return &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.bssCluster.Name + "-config",
            Namespace: b.bssCluster.Namespace,
            Labels:    CommonLabels(b.bssCluster),
        },
        Data: map[string]string{
            "app.conf": "# Your config here",
        },
    }
}
```

#### Step 2: Create the Resource Reconciler

Create `internal/resources/configmap.go`:

```go
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

type ConfigMapReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func NewConfigMapReconciler(c client.Client, scheme *runtime.Scheme) *ConfigMapReconciler {
    return &ConfigMapReconciler{Client: c, Scheme: scheme}
}

func (r *ConfigMapReconciler) Reconcile(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
    desired := builder.NewConfigMapBuilder(bssCluster).Build()
    
    existing := &corev1.ConfigMap{}
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

func (r *ConfigMapReconciler) create(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, cm *corev1.ConfigMap, log logr.Logger) error {
    if err := controllerutil.SetControllerReference(bssCluster, cm, r.Scheme); err != nil {
        return err
    }
    
    log.Info("Creating ConfigMap", "name", cm.Name)
    return r.Create(ctx, cm)
}

func (r *ConfigMapReconciler) update(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, existing, desired *corev1.ConfigMap, log logr.Logger) error {
    desired.ResourceVersion = existing.ResourceVersion
    
    if err := controllerutil.SetControllerReference(bssCluster, desired, r.Scheme); err != nil {
        return err
    }
    
    // Compare and update only if needed
    if r.needsUpdate(existing, desired) {
        log.Info("Updating ConfigMap", "name", desired.Name)
        return r.Update(ctx, desired)
    }
    
    return nil
}

func (r *ConfigMapReconciler) needsUpdate(existing, desired *corev1.ConfigMap) bool {
    // Compare data
    if len(existing.Data) != len(desired.Data) {
        return true
    }
    for k, v := range desired.Data {
        if existing.Data[k] != v {
            return true
        }
    }
    return false
}
```

#### Step 3: Add to Main Controller

In `internal/controller/bsscluster_controller.go`:

```go
// Add to struct
type BssClusterReconciler struct {
    client.Client
    Scheme *runtime.Scheme

    statefulSetReconciler *resources.StatefulSetReconciler
    serviceReconciler     *resources.ServiceReconciler
    configMapReconciler   *resources.ConfigMapReconciler  // ADD THIS

    validator *validation.Validator
}

// Update constructor
func NewBssClusterReconciler(c client.Client, scheme *runtime.Scheme) *BssClusterReconciler {
    return &BssClusterReconciler{
        Client:                c,
        Scheme:                scheme,
        statefulSetReconciler: resources.NewStatefulSetReconciler(c, scheme),
        serviceReconciler:     resources.NewServiceReconciler(c, scheme),
        configMapReconciler:   resources.NewConfigMapReconciler(c, scheme),  // ADD THIS
        validator:             validation.NewValidator(),
    }
}

// Add to reconcileResources
func (r *BssClusterReconciler) reconcileResources(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
    // Reconcile ConfigMap first
    if err := r.configMapReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
    
    // ... rest of resources
}
```

#### Step 4: Add RBAC Markers

At the top of `bsscluster_controller.go`, add:

```go
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
```

#### Step 5: Run Code Generation

```bash
make manifests
make generate
```

## Resource Reconciliation Order

Resources should be reconciled in dependency order:

1. **ConfigMaps** - Configuration data
2. **Secrets** - Sensitive data
3. **PVCs** - Persistent volumes (if not using StatefulSet VCTs)
4. **Service** - Required before StatefulSet for stable network identity
5. **StatefulSet** - The main workload
6. **Ingress** - External access (after Service exists)

## Testing Pattern

When adding tests for new resources:

```go
It("should create ConfigMap", func() {
    // ... test implementation
    configMap := &corev1.ConfigMap{}
    Eventually(func() error {
        return k8sClient.Get(ctx, types.NamespacedName{
            Name: resourceName + "-config",
            Namespace: "default",
        }, configMap)
    }, timeout, interval).Should(Succeed())
    
    Expect(configMap.Data).To(HaveKey("app.conf"))
})
```

## Best Practices

### 1. Immutable Fields
Always preserve immutable fields when updating:
```go
desired.ResourceVersion = existing.ResourceVersion
desired.Spec.ClusterIP = existing.Spec.ClusterIP  // For Services
```

### 2. Owner References
Always set owner references for garbage collection:
```go
controllerutil.SetControllerReference(bssCluster, resource, r.Scheme)
```

### 3. Conditional Reconciliation
Only reconcile resources when needed:
```go
if bssCluster.Spec.EnableIngress {
    if err := r.ingressReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
}
```

### 4. Status Updates
Update status to reflect reality:
```go
bssCluster.Status.ConfigMapReady = true
bssCluster.Status.Conditions = []metav1.Condition{...}
```

## Common Resources to Add

- **ConfigMap**: Application configuration
- **Secret**: Sensitive data (passwords, certificates)
- **PVC**: Persistent storage (alternative to StatefulSet VCTs)
- **Ingress**: External HTTP/HTTPS access
- **HPA**: Horizontal Pod Autoscaler
- **NetworkPolicy**: Network segmentation
- **ServiceAccount**: Pod identity
- **RBAC**: Roles and RoleBindings for the workload

## Debugging

Enable verbose logging:
```go
log.V(1).Info("Detailed debug message", "key", value)
```

Check resource events:
```bash
kubectl describe bsscluster <name>
kubectl get events --sort-by='.lastTimestamp'
```

## Further Reading

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [controller-runtime Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
