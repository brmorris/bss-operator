# BSS Operator Architecture Diagram

## Component Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Kubernetes API                                  │
│                       (Watches BssCluster CRD events)                       │
└────────────────────────────────────┬────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     Controller (Orchestrator)                                │
│                  internal/controller/bsscluster_controller.go                │
│                                                                               │
│  Reconcile(ctx, req):                                                        │
│    1. Fetch BssCluster                                                       │
│    2. Validate Spec          ◄────────┐                                     │
│    3. Update Status (Reconciling)     │                                     │
│    4. Reconcile Resources     ────────┼──────┐                              │
│    5. Update Status (Ready/Failed)    │      │                              │
└───────────────────────────────────────┼──────┼──────────────────────────────┘
                                        │      │
                  ┌─────────────────────┘      └─────────────────────┐
                  │                                                   │
                  ▼                                                   ▼
    ┌──────────────────────────┐                      ┌──────────────────────────┐
    │  Validation Package      │                      │  Resources Package       │
    │  internal/validation/    │                      │  internal/resources/     │
    │                          │                      │                          │
    │  • validator.go          │                      │  Resource Reconcilers:   │
    │    - Validate()          │                      │  • statefulset.go        │
    │    - validateSpec()      │                      │  • service.go            │
    │                          │                      │  • configmap.go (stub)   │
    └──────────────────────────┘                      │  • secret.go (stub)      │
                                                       │  • ingress.go (stub)     │
                                                       │  • pvc.go (stub)         │
                                                       │                          │
                                                       │  Each reconciler:        │
                                                       │  • Reconcile()           │
                                                       │  • create()              │
                                                       │  • update()              │
                                                       │  • needsUpdate()         │
                                                       └──────────┬───────────────┘
                                                                  │
                                                                  │ Uses
                                                                  ▼
                                                  ┌──────────────────────────────┐
                                                  │  Builder Package             │
                                                  │  internal/builder/           │
                                                  │                              │
                                                  │  Pure Functions:             │
                                                  │  • labels.go                 │
                                                  │    - CommonLabels()          │
                                                  │    - SelectorLabels()        │
                                                  │                              │
                                                  │  • statefulset_builder.go    │
                                                  │    - Build()                 │
                                                  │                              │
                                                  │  • service_builder.go        │
                                                  │    - Build()                 │
                                                  │                              │
                                                  │  Outputs: Desired K8s Objects│
                                                  └──────────────────────────────┘
```

## Reconciliation Flow (Detailed)

```
User Creates/Updates BssCluster
         │
         ▼
┌────────────────────────┐
│   Controller Watch     │
│   Event Triggered      │
└───────────┬────────────┘
            │
            ▼
┌────────────────────────┐
│ Fetch BssCluster CR    │
└───────────┬────────────┘
            │
            ▼
┌────────────────────────┐
│ Validate Spec          │◄──── Validator.Validate()
└───────────┬────────────┘
            │
            ▼
┌────────────────────────┐
│ Status = Reconciling   │
└───────────┬────────────┘
            │
            ▼
┌─────────────────────────────────────────┐
│    Reconcile Resources (in order)       │
├─────────────────────────────────────────┤
│  1. Service Reconciler                  │
│     ├─ ServiceBuilder.Build()           │
│     ├─ Get existing Service             │
│     └─ Create or Update                 │
│                                         │
│  2. StatefulSet Reconciler              │
│     ├─ StatefulSetBuilder.Build()       │
│     ├─ Get existing StatefulSet         │
│     └─ Create or Update                 │
│                                         │
│  3. [Future: ConfigMap Reconciler]      │
│  4. [Future: Secret Reconciler]         │
│  5. [Future: Ingress Reconciler]        │
└────────────┬────────────────────────────┘
             │
             ▼
    ┌────────────────────┐
    │ All Resources OK?  │
    └────────┬───────────┘
             │
        ┌────┴────┐
        │         │
       Yes       No
        │         │
        ▼         ▼
  ┌─────────┐  ┌──────────┐
  │ Status  │  │  Status  │
  │ = Ready │  │ = Failed │
  └─────────┘  └──────────┘
```

## Resource Reconciler Pattern (Each Resource)

```
┌────────────────────────────────────────┐
│   Reconcile(ctx, bssCluster, log)      │
└─────────────────┬──────────────────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Build Desired      │◄──── Builder.Build()
         │ State              │
         └────────┬───────────┘
                  │
                  ▼
         ┌────────────────────┐
         │ Get Existing       │
         │ Resource           │
         └────────┬───────────┘
                  │
            ┌─────┴──────┐
            │            │
         Exists      Not Found
            │            │
            ▼            ▼
    ┌──────────────┐  ┌──────────────┐
    │   update()   │  │   create()   │
    │              │  │              │
    │ • Preserve   │  │ • Set Owner  │
    │   immutable  │  │   Reference  │
    │ • Set owner  │  │ • Log action │
    │ • Compare    │  │ • Create     │
    │ • Update if  │  │              │
    │   needed     │  │              │
    └──────────────┘  └──────────────┘
```

## Data Flow

```
BssCluster CRD
      │
      ▼
┌──────────────┐
│  Validation  │
└──────┬───────┘
       │ Valid Spec
       ▼
┌──────────────┐
│   Builder    │ ──────► Desired State (K8s Object)
└──────────────┘
       │
       ▼
┌──────────────┐
│  Reconciler  │ ──────► Existing State
└──────┬───────┘
       │
       ▼
   Compare & Apply
       │
       ▼
   Kubernetes API
       │
       ▼
   Actual Resources
```

## Adding New Resource (Flow)

```
Developer wants to add Ingress support
         │
         ▼
┌─────────────────────────────────┐
│ 1. Create IngressBuilder        │
│    internal/builder/             │
│    ingress_builder.go            │
│    • Build() method              │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ 2. Create IngressReconciler     │
│    internal/resources/           │
│    ingress.go                    │
│    • Reconcile() method          │
│    • create(), update()          │
│    • needsUpdate()               │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ 3. Wire into Controller         │
│    • Add reconciler field        │
│    • Initialize in constructor   │
│    • Call in reconcileResources  │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ 4. Add RBAC Marker              │
│    // +kubebuilder:rbac:...     │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ 5. Generate & Test              │
│    make manifests generate test │
└─────────────────────────────────┘
```

## Key Principles

1. **Single Direction Flow**: Controller → Reconcilers → Builders
2. **No Circular Dependencies**: Builders never call reconcilers
3. **Pure Functions**: Builders have no side effects
4. **Idempotency**: Reconcilers can be called repeatedly safely
5. **Owner References**: Automatic cleanup via K8s garbage collection

## Testing Boundaries

```
Unit Tests          Integration Tests        E2E Tests
    │                      │                      │
    ▼                      ▼                      ▼
┌─────────┐         ┌──────────┐          ┌──────────┐
│Builders │         │Reconciler│          │Controller│
│         │         │ + envtest│          │ + Kind   │
│Pure Fns │         │          │          │          │
└─────────┘         └──────────┘          └──────────┘
```

## File Organization

```
internal/
├── controller/
│   ├── bsscluster_controller.go       ← Orchestrator (80 lines)
│   ├── bsscluster_controller_test.go  ← Integration tests
│   └── suite_test.go
│
├── resources/                          ← One file per K8s resource
│   ├── statefulset.go                 ← ~150 lines each
│   ├── service.go
│   ├── configmap.go (stub)
│   ├── secret.go (stub)
│   ├── ingress.go (stub)
│   └── pvc.go (stub)
│
├── builder/                            ← Pure construction functions
│   ├── labels.go                      ← ~50 lines
│   ├── statefulset_builder.go        ← ~100 lines
│   └── service_builder.go             ← ~60 lines
│
└── validation/                         ← Validation logic
    └── validator.go                    ← ~50 lines
```

---

**Total Lines of Code**: ~500 lines (was 250 monolithic)
**But**: Infinitely more scalable and maintainable!
