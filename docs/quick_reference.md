# Quick Reference: Adding New Resources

## TL;DR - 5 Steps to Add a Resource

1. **Create Builder** → `internal/builder/<resource>_builder.go`
2. **Create Reconciler** → `internal/resources/<resource>.go`
3. **Wire into Controller** → Add to `BssClusterReconciler`
4. **Add RBAC** → Add kubebuilder marker
5. **Run Codegen** → `make manifests generate`

---

## Example: Adding Ingress Support in 5 Minutes

### 1. Builder (`internal/builder/ingress_builder.go`)

```go
package builder

import (
    networkingv1 "k8s.io/api/networking/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
)

type IngressBuilder struct {
    bssCluster *bssv1alpha1.BssCluster
}

func NewIngressBuilder(bssCluster *bssv1alpha1.BssCluster) *IngressBuilder {
    return &IngressBuilder{bssCluster: bssCluster}
}

func (b *IngressBuilder) Build() *networkingv1.Ingress {
    pathType := networkingv1.PathTypePrefix
    return &networkingv1.Ingress{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.bssCluster.Name,
            Namespace: b.bssCluster.Namespace,
            Labels:    CommonLabels(b.bssCluster),
        },
        Spec: networkingv1.IngressSpec{
            Rules: []networkingv1.IngressRule{{
                Host: b.bssCluster.Name + ".example.com",
                IngressRuleValue: networkingv1.IngressRuleValue{
                    HTTP: &networkingv1.HTTPIngressRuleValue{
                        Paths: []networkingv1.HTTPIngressPath{{
                            Path:     "/",
                            PathType: &pathType,
                            Backend: networkingv1.IngressBackend{
                                Service: &networkingv1.IngressServiceBackend{
                                    Name: b.bssCluster.Name,
                                    Port: networkingv1.ServiceBackendPort{
                                        Number: 8080,
                                    },
                                },
                            },
                        }},
                    },
                },
            }},
        },
    }
}
```

### 2. Reconciler (`internal/resources/ingress.go`)

```go
package resources

import (
    "context"
    "github.com/go-logr/logr"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

    bssv1alpha1 "github.com/brmorris/bss-operator/api/v1alpha1"
    "github.com/brmorris/bss-operator/internal/builder"
)

type IngressReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func NewIngressReconciler(c client.Client, scheme *runtime.Scheme) *IngressReconciler {
    return &IngressReconciler{Client: c, Scheme: scheme}
}

func (r *IngressReconciler) Reconcile(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
    desired := builder.NewIngressBuilder(bssCluster).Build()
    existing := &networkingv1.Ingress{}

    err := r.Get(ctx, types.NamespacedName{
        Name: desired.Name, Namespace: desired.Namespace,
    }, existing)

    if err != nil {
        if errors.IsNotFound(err) {
            if err := controllerutil.SetControllerReference(bssCluster, desired, r.Scheme); err != nil {
                return err
            }
            log.Info("Creating Ingress", "name", desired.Name)
            return r.Create(ctx, desired)
        }
        return err
    }

    // Update if needed (simplified - add comparison logic)
    return nil
}
```

### 3. Wire into Controller

In `internal/controller/bsscluster_controller.go`:

```go
// Add field
type BssClusterReconciler struct {
    // ... existing fields ...
    ingressReconciler *resources.IngressReconciler  // ADD
}

// Update constructor
func NewBssClusterReconciler(c client.Client, scheme *runtime.Scheme) *BssClusterReconciler {
    return &BssClusterReconciler{
        // ... existing ...
        ingressReconciler: resources.NewIngressReconciler(c, scheme),  // ADD
    }
}

// Add to reconcileResources (at the END, after Service)
func (r *BssClusterReconciler) reconcileResources(...) error {
    // ... existing resources ...

    // Reconcile Ingress (optional, could be conditional)
    if err := r.ingressReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }

    return nil
}
```

### 4. Add RBAC

At top of `bsscluster_controller.go`:

```go
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
```

### 5. Generate Code

```bash
make manifests generate
make test
```

---

## Resource Order Matters

Reconcile in this order:

1. ConfigMaps
2. Secrets
3. Service (before StatefulSet!)
4. StatefulSet
5. Ingress

---

## Common Patterns

### Conditional Resources

```go
// Only create if enabled in spec
if bssCluster.Spec.Ingress != nil && bssCluster.Spec.Ingress.Enabled {
    if err := r.ingressReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
}
```

### Multiple Resources of Same Type

```go
// Use different builders for public/private services
publicSvc := builder.NewPublicServiceBuilder(bssCluster).Build()
privateSvc := builder.NewPrivateServiceBuilder(bssCluster).Build()
```

### Dynamic Resource Names

```go
// In builder
Name: fmt.Sprintf("%s-%s", b.bssCluster.Name, "monitoring"),
```

---

## Testing New Resources

Add to `bsscluster_controller_test.go`:

```go
It("should create Ingress", func() {
    ingress := &networkingv1.Ingress{}
    Eventually(func() error {
        return k8sClient.Get(ctx, typeNamespacedName, ingress)
    }, timeout, interval).Should(Succeed())

    Expect(ingress.Spec.Rules).ToNot(BeEmpty())
})
```

---

## Debugging

```bash
# Check what controller is doing
kubectl logs -n bss-operator-system deployment/bss-operator-controller-manager -f

# See resource events
kubectl describe bsscluster <name>

# Check generated RBAC
cat config/rbac/role.yaml
```

---

## Full Documentation

See `docs/controller_architecture.md` for complete guide with all patterns and best practices.
