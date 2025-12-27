# Controller Summary

```
internal/
├── controller/          # Orchestration (80 lines)
│   └── bsscluster_controller.go
├── resources/          # Resource reconcilers
│   ├── statefulset.go
│   ├── service.go
│   ├── configmap.go   (stub)
│   ├── secret.go      (stub)
│   ├── ingress.go     (stub)
│   └── pvc.go         (stub)
├── builder/           # Pure construction functions
│   ├── labels.go
│   ├── statefulset_builder.go
│   └── service_builder.go
└── validation/        # Validation logic
    └── validator.go
```

## Themes

### 1. **Separation of Concerns**
- **Controller**: Orchestrates reconciliation
- **Reconcilers**: Handle resource CRUD
- **Builders**: Construct K8s objects
- **Validators**: Check specifications

### 2. **Single Responsibility Principle**
Each reconciler manages exactly one Kubernetes resource type.

### 3. **Scalability**
Adding a new resource now requires:
- 1 builder file (~50 lines)
- 1 reconciler file (~100 lines)
- 3 lines in controller
- 1 RBAC marker

### 4. **Testability**
- **Builders**: Unit testable (pure functions)
- **Reconcilers**: Integration testable
- **Controller**: E2E testable

### 5. **Maintainability**
Clear structure makes code navigation and debugging easier.

## 📁 Structure

### Core Components

#### **Builder Pattern**
```go
// Pure function - no side effects
func (b *ServiceBuilder) Build() *corev1.Service {
    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:   b.bssCluster.Name,
            Labels: CommonLabels(b.bssCluster),
        },
        Spec: // ... service spec
    }
}
```

#### **Reconciler Pattern**
```go
// Manages one resource type
type ServiceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *ServiceReconciler) Reconcile(ctx context.Context,
    bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
    // Get desired state from builder
    desired := builder.NewServiceBuilder(bssCluster).Build()

    // Compare with existing
    // Create or update as needed
}
```

#### **Controller Pattern**
```go
// Orchestrates all reconcilers
func (r *BssClusterReconciler) reconcileResources(...) error {
    // Reconcile in dependency order
    if err := r.serviceReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
    if err := r.statefulSetReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
    // Easy to add more!
    return nil
}
```

## Example Usage

### Adding a New Resource (ConfigMap Example)

1. **Create builder** (`internal/builder/configmap_builder.go`):
```go
func (b *ConfigMapBuilder) Build() *corev1.ConfigMap {
    return &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.bssCluster.Name + "-config",
            Namespace: b.bssCluster.Namespace,
            Labels:    CommonLabels(b.bssCluster),
        },
        Data: map[string]string{
            "app.conf": "# Your config",
        },
    }
}
```

2. **Create reconciler** (`internal/resources/configmap.go`):
   - Copy pattern from `statefulset.go`
   - Adjust for ConfigMap type
   - Implement comparison logic

3. **Wire into controller**:
```go
// Add field
configMapReconciler *resources.ConfigMapReconciler

// Initialize in constructor
configMapReconciler: resources.NewConfigMapReconciler(c, scheme),

// Call in reconcileResources
if err := r.configMapReconciler.Reconcile(ctx, bssCluster, log); err != nil {
    return err
}
```

4. **Add RBAC marker**:
```go
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
```

5. **Generate and test**:
```bash
make manifests generate test
```

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| `docs/controller_architecture.md` | Complete implementation guide with all patterns |
| `internal/README.md` | Package structure and organization |

## ✅ Tests Pass

```bash
$ make test
ok  github.com/brmorris/bss-operator/internal/controller  9.895s  coverage: 44.4%
```

All existing functionality preserved, tests updated and passing.

## 🎓 Learning Path

This structure is **perfect for learning** because:

1. **Clear Separation**: Each concept (building, reconciling, validating) is isolated
2. **Easy Experimentation**: Add resources without breaking existing code
3. **Real-World Pattern**: Used in production operators (Prometheus, Istio, etc.)
4. **Incremental Learning**: Start simple, add complexity as you learn

### Suggested Learning Order

1. ✅ **Done**: Understand current StatefulSet + Service
2. **Next**: Add ConfigMap (simplest resource)
3. **Then**: Add Secret (similar to ConfigMap)
4. **Advanced**: Add Ingress (more complex spec)
5. **Expert**: Add PVCs, HPA, NetworkPolicy

## 🔧 Maintenance

### When to Split Further

If a reconciler exceeds ~200 lines:
- Extract complex comparison logic
- Create specialized sub-reconcilers
- Add helper functions

### When to Add Packages

As you grow:
- `internal/status/` - Complex status management
- `internal/conditions/` - Status condition helpers
- `internal/events/` - Event recording
- `internal/metrics/` - Custom metrics

## 🎯 Production Patterns Included

- ✅ Owner references for garbage collection
- ✅ Immutable field preservation
- ✅ Comparison before update (efficiency)
- ✅ Structured logging
- ✅ Error handling patterns
- ✅ RBAC management
- ✅ Resource ordering

## 🔍 Next Steps

1. **Experiment**: Try adding ConfigMap following the docs
2. **Extend CRD**: Add fields to BssClusterSpec as you need them
3. **Add Conditions**: Implement status conditions for better observability
4. **Learn Finalizers**: When you need cleanup logic
5. **Advanced**: Webhooks for validation and mutation

## 💡 Pro Tips

1. Always reconcile in dependency order (ConfigMap → Service → StatefulSet)
2. Use `make manifests generate` after any API changes
3. Check `config/rbac/role.yaml` to see generated permissions
4. Enable verbose logging: `log.V(1).Info("debug message")`
5. Test locally with kind before deploying

## 🆘 Getting Help

- Read `docs/controller_architecture.md` for detailed patterns
- Check `docs/quick_reference.md` for quick answers
- Look at existing reconcilers as examples
- Kubebuilder Book: https://book.kubebuilder.io/

---

**Your operator is now ready to scale!** 🚀

Start by adding a ConfigMap or Secret to practice the pattern, then move on to more complex resources as you learn.
