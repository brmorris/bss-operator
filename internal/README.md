# Internal Packages

This directory contains the core implementation of the BSS Operator, organized into a scalable, enterprise-grade architecture.

## Package Overview

### 📦 `controller/`
Main reconciliation loop and orchestration logic.

- **Purpose**: Coordinates resource reconciliation
- **Key Files**: 
  - `bsscluster_controller.go` - Main reconciler
  - `bsscluster_controller_test.go` - Integration tests

### 📦 `resources/`
Resource-specific reconcilers - one per Kubernetes resource type.

- **Purpose**: Handle CRUD operations for specific resources
- **Pattern**: Each reconciler manages exactly one K8s resource type
- **Implemented**: 
  - `statefulset.go` ✅
  - `service.go` ✅
- **Ready to implement**:
  - `configmap.go` 📝
  - `secret.go` 📝
  - `ingress.go` 📝
  - `pvc.go` 📝

### 📦 `builder/`
Pure functions that construct Kubernetes objects.

- **Purpose**: Build desired state for resources
- **Pattern**: Pure functions with no side effects
- **Key Files**:
  - `labels.go` - Common label management
  - `statefulset_builder.go` - StatefulSet construction
  - `service_builder.go` - Service construction

### 📦 `validation/`
Validation logic for custom resources.

- **Purpose**: Validate BssCluster specifications
- **Key Files**:
  - `validator.go` - Main validation logic

## Architecture Benefits

### ✅ **Separation of Concerns**
- Controller orchestrates
- Reconcilers manage resources
- Builders construct objects
- Validators check specs

### ✅ **Single Responsibility**
Each package and file has one clear purpose.

### ✅ **Testability**
- Builders: Unit testable (pure functions)
- Reconcilers: Integration testable
- Controller: End-to-end testable

### ✅ **Scalability**
Adding new resources is straightforward and doesn't bloat existing files.

### ✅ **Maintainability**
Clear structure makes code easy to navigate and modify.

## Adding New Resources

See the detailed guides:
- **Full Guide**: `../docs/controller_architecture.md`
- **Quick Reference**: `../docs/quick_reference.md`

### Quick Example

To add a ConfigMap:

1. Create `builder/configmap_builder.go`
2. Create `resources/configmap.go`
3. Wire into `controller/bsscluster_controller.go`
4. Add RBAC marker
5. Run `make manifests generate test`

## File Organization Rules

### Builders (`builder/`)
- **Naming**: `<resource>_builder.go`
- **Content**: Pure construction functions
- **Dependencies**: Only K8s API types and our CRDs
- **No**: Client calls, logging, context

### Reconcilers (`resources/`)
- **Naming**: `<resource>.go`
- **Content**: CRUD logic for one resource type
- **Pattern**: Reconcile(), create(), update(), needsUpdate()
- **Dependencies**: Client, Scheme, Builders

### Controller (`controller/`)
- **Content**: Orchestration only
- **Delegates**: All resource management to reconcilers
- **Owns**: Resource ordering and status updates

## Code Standards

```go
// ✅ Good: Reconciler pattern
func (r *ServiceReconciler) Reconcile(ctx context.Context, bssCluster *bssv1alpha1.BssCluster, log logr.Logger) error {
    desired := builder.NewServiceBuilder(bssCluster).Build()
    // ... reconciliation logic
}

// ✅ Good: Builder pattern
func (b *ServiceBuilder) Build() *corev1.Service {
    return &corev1.Service{
        // ... pure construction
    }
}

// ❌ Bad: Mixing concerns
func (r *ServiceReconciler) BuildAndReconcile() error {
    // Don't mix building and reconciling
}
```

## Testing Strategy

- **Unit Tests**: Builders (pure functions)
- **Integration Tests**: Reconcilers (with envtest)
- **E2E Tests**: Controller (full cluster)

## Performance Considerations

- **Lazy Reconciliation**: Only update when necessary
- **Comparison Logic**: Check `needsUpdate()` before API calls
- **Owner References**: Let K8s handle cascading deletes
- **Watch Events**: Controller-runtime handles efficiently

## Common Patterns

### Resource Creation
```go
if err := controllerutil.SetControllerReference(bssCluster, resource, r.Scheme); err != nil {
    return err
}
log.Info("Creating Resource", "name", resource.Name)
return r.Create(ctx, resource)
```

### Resource Updates
```go
desired.ResourceVersion = existing.ResourceVersion
if r.needsUpdate(existing, desired) {
    log.Info("Updating Resource", "name", desired.Name)
    return r.Update(ctx, desired)
}
```

### Conditional Logic
```go
if bssCluster.Spec.EnableFeature {
    if err := r.featureReconciler.Reconcile(ctx, bssCluster, log); err != nil {
        return err
    }
}
```

## Troubleshooting

### "Resource not found" errors
- Check RBAC markers are present
- Run `make manifests` to regenerate
- Check `config/rbac/role.yaml`

### Tests failing
- Ensure resources reconciled in correct order
- Check owner references are set
- Verify test timeouts are sufficient

### Updates not applying
- Check `needsUpdate()` comparison logic
- Verify ResourceVersion is preserved
- Look for immutable field conflicts

## Resources

- 📚 [Controller Architecture Guide](../docs/controller_architecture.md)
- ⚡ [Quick Reference](../docs/quick_reference.md)
- 📖 [Kubebuilder Book](https://book.kubebuilder.io/)
- 🔧 [controller-runtime Docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
