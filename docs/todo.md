# Mini bss Operator — Actionable TODO List

**Goal:**
Learn Humio / bss operator patterns by building a toy operator using **Operator SDK + Kubebuilder (Go)**.

---

## Week 1 — Core Operator Fundamentals

---

### ⬜ 1. Required tooling

* Install **Go** (≥ 1.22)
* Install **Operator SDK**
* Install a local Kubernetes cluster (**kind** or **minikube**)

---

### ⬜ 2. Initialize the operator project

Verify the following exist:

* `PROJECT`
* `main.go`
* `config/`
* `api/`
* `controllers/`

> **Note:**
> The `PROJECT` file should reference a **kubebuilder layout**.
> This matches the Humio / bss operator structure.

---

### ⬜ 3. Create `BssCluster` API + controller

```bash
operator-sdk create api \
  --group bss \
  --version v1alpha1 \
  --kind BssCluster \
  --resource --controller
```

Edit:

```
api/v1alpha1/BssCluster_types.go
```

Add spec and status:

```go
type BssClusterSpec struct {
    Replicas *int32 `json:"replicas,omitempty"`
    Image    string `json:"image"`
}

type BssClusterStatus struct {
    Phase string `json:"phase,omitempty"`
}
```

Generate CRDs:

```bash
make generate
make manifests
```

---

### ⬜ 4. Implement the basic Reconcile loop

File:

```
controllers/BssCluster_controller.go
```

Add the reconciler skeleton:

```go
func (r *BssClusterReconciler) Reconcile(
    ctx context.Context,
    req ctrl.Request,
) (ctrl.Result, error) {

    var cluster bssv1alpha1.BssCluster
    if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    cluster.Status.Phase = "Reconciling"
    _ = r.Status().Update(ctx, &cluster)

    // TODO brad: create or update StatefulSet
    // TODO: create or update Service

    cluster.Status.Phase = "Ready"
    _ = r.Status().Update(ctx, &cluster)

    return ctrl.Result{}, nil
}
```

**Concept learned:**
Reconciliation = converge actual state toward desired state.

---

### ⬜ 5. Create a StatefulSet for the cluster

Add a helper function:

```go
func desiredStatefulSet(
    cluster *bssv1alpha1.BssCluster,
) *appsv1.StatefulSet {

    replicas := int32(1)
    if cluster.Spec.Replicas != nil {
        replicas = *cluster.Spec.Replicas
    }

    return &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      cluster.Name,
            Namespace: cluster.Namespace,
        },
        Spec: appsv1.StatefulSetSpec{
            Replicas: &replicas,
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "bss",
                            Image: cluster.Spec.Image,
                        },
                    },
                },
            },
        },
    }
}
```

In `Reconcile`:

* Fetch existing StatefulSet
* Create if missing
* Update if spec differs

> This closely mirrors how the Humio operator manages bss nodes.

---

### ⬜ 6. Add a finalizer for cleanup

Define a finalizer string:

```
bss.example.com/finalizer
```

Add finalizer logic:

```go
if !controllerutil.ContainsFinalizer(
    &cluster,
    "bss.example.com/finalizer",
) {
    controllerutil.AddFinalizer(
        &cluster,
        "bss.example.com/finalizer",
    )
    r.Update(ctx, &cluster)
}

if !cluster.ObjectMeta.DeletionTimestamp.IsZero() {
    // cleanup dependent resources here
    controllerutil.RemoveFinalizer(
        &cluster,
        "bss.example.com/finalizer",
    )
    r.Update(ctx, &cluster)
    return ctrl.Result{}, nil
}
```

**Concept learned:**
Safe deletion and lifecycle control (used heavily in Humio).

---

## Week 2 — Humio-Like Operator Patterns

---

### ⬜ 7. Add `MiniIngestToken` CRD

```bash
operator-sdk create api \
  --group bss \
  --version v1alpha1 \
  --kind MiniIngestToken \
  --resource --controller
```

Example CR:

```yaml
apiVersion: bss.example.com/v1alpha1
kind: MiniIngestToken
metadata:
  name: demo-token
spec:
  clusterRef: demo
  permissions:
    - ingest
```

---

### ⬜ 8. Implement `MiniIngestToken` reconciler

Responsibilities:

* Verify referenced cluster exists
* Generate a fake token
* Store token in a Secret
* Update status

Example logic:

```go
token := uuid.NewString()

secret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      tokenCR.Name,
        Namespace: tokenCR.Namespace,
    },
    StringData: map[string]string{
        "token": token,
    },
}
```

**Concept learned:**
Cross-CR dependencies and secret lifecycle management.

---

### ⬜ 9. Replace `Phase` with status conditions

Update status definition:

```go
type BssClusterStatus struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

Use:

```go
meta.SetStatusCondition(&cluster.Status.Conditions, metav1.Condition{
    Type:   "Ready",
    Status: metav1.ConditionTrue,
    Reason: "Reconciled",
})
```

**Concept learned:**
Production-grade operator observability.

---

### ⬜ 10. Watch dependent resources

Update controller setup to watch:

* `StatefulSet`
* `Service`
* `Secret`

**Concept learned:**
Reconciliation triggered by downstream resource changes.

---

### ⬜ 11. Add validation & defaulting webhooks (optional)

```bash
operator-sdk create webhook \
  --group bss \
  --version v1alpha1 \
  --kind BssCluster \
  --defaulting --validation
```

Validation rules:

* `replicas >= 1`
* `image` must not be empty

**Concept learned:**
API safety and early failure prevention.

---

### ⬜ 12. Deploy and test the operator locally

```bash
make install
make run
```

Apply a cluster:

```yaml
apiVersion: bss.example.com/v1alpha1
kind: BssCluster
metadata:
  name: demo
spec:
  replicas: 3
  image: nginx
```

Verify:

* StatefulSet created
* Status updated
* Finalizer present
* Reconcile loops fire as expected

---

## Final Checkpoint

⬜ Read the **Humio / bss operator code** and map:

* CRDs → your CRDs
* Reconcilers → your reconcilers
* Finalizers → your cleanup logic
