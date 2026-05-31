# manifestation-operator

> Your pods are not broken. They are simply *awaiting your intention.*

A Kubernetes operator for the manifestation generation. Pods do not schedule because
some YAML told them to. Pods schedule because **you believe**, out loud, in present
tense, that they already are.

```
NAMESPACE      NAME                     READY   STATUS            REASON
manifest-demo  nginx-7c9d-abcde         0/1     Pending           SchedulingGated
```

That pod is not stuck. It is in **limbo**, waiting for you to speak your truth. And no,
checking `kubectl describe` for the fifth time will not help. It can tell you are anxious.

```yaml
apiVersion: manifestation.pez.sh/v1alpha1
kind: Desire
metadata:
  name: nginx-is-serving
  namespace: manifest-demo
spec:
  manifestation: "My nginx pod is healthy and serving traffic"
  intensity: chant
```

You have to really believe it, though. The scheduler can tell.

```
NAMESPACE      NAME                     READY   STATUS    REASON
manifest-demo  nginx-7c9d-abcde         1/1     Running
```

You did that. With your mind.

## How it works

There is real engineering under the joke.

1. A **mutating admission webhook** intercepts every pod created in the cluster (except
   control-plane namespaces like `kube-system`) and injects a
   [scheduling gate](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-scheduling-readiness/)
   named `manifestation.pez.sh/awaiting-manifestation`. The scheduler refuses to bind a
   gated pod, so it sits `Pending` / `SchedulingGated` indefinitely. No CPU burned, no
   crash loop - just patience.

2. You create a **`Desire`** custom resource. The controller judges your
   `spec.manifestation`. It **must be present tense**, stated as already true. Future
   tense, conditionals, hope, wishing, questions, and dwelling on the past are all
   rejected - the universe does not respond to doubt.

3. If the affirmation is worthy, the controller **removes the scheduling gate** from
   every matching pod in the namespace and annotates it with
   `manifestation.pez.sh/manifested-by`. The scheduler does the rest. Your desire
   becomes reality.

A standing `Desire` keeps working: pods created later are gated by the webhook and then
released by the controller (it watches pods and re-reconciles), so your intention
persists.

### What the universe rejects

| You wrote | Why it fails |
|-----------|--------------|
| `My pod will be healthy` | future tense - `will` |
| `The pod should be running` | conditional - `should` |
| `I hope the database is up` | wishing - `hope` |
| `Is my pod healthy?` | a question is doubt wearing punctuation |
| `The pod was healthy` | dwelling on the past |
| `Please make it work` | `please` betrays need |

### What the universe accepts

- `My nginx pod is healthy and serving traffic`
- `The database is running and replication flows`
- `Everything is fine` *(it is)*

`spec.intensity` (`whisper` / `speak` / `shout` / `chant`) does not change physics.
It changes you. Studies (none) show `chant` users report 40% more alignment with their
infrastructure.

## Quick start

Requires a cluster running Kubernetes **v1.27+** (scheduling gates) and
[cert-manager](https://cert-manager.io/) for the webhook's TLS.

### Install with Helm

The chart is published as an OCI artifact on Docker Hub:

```sh
helm install manifestation-operator \
  oci://registry-1.docker.io/rwejlgaard/manifestation-operator-chart \
  --namespace manifestation-operator-system --create-namespace
```

Common overrides:

```sh
# Pin a version (recommended)
helm install manifestation-operator \
  oci://registry-1.docker.io/rwejlgaard/manifestation-operator-chart \
  --version 1.0.0 \
  --namespace manifestation-operator-system --create-namespace

# Bring your own webhook cert instead of cert-manager
helm install manifestation-operator \
  oci://registry-1.docker.io/rwejlgaard/manifestation-operator-chart \
  --set webhook.certManager.enabled=false \
  --set webhook.certSecretName=my-webhook-cert \
  --namespace manifestation-operator-system --create-namespace
```

See [`charts/manifestation-operator/values.yaml`](charts/manifestation-operator/values.yaml)
for the full set of knobs (webhook scope, metrics, RBAC, resources). To upgrade or remove:

```sh
helm upgrade manifestation-operator \
  oci://registry-1.docker.io/rwejlgaard/manifestation-operator-chart \
  -n manifestation-operator-system

helm uninstall manifestation-operator -n manifestation-operator-system
```

`helm uninstall` leaves the CRD in place (`helm.sh/resource-policy: keep`) so your
`Desire` objects survive. Delete it by hand if you truly want to let go:
`kubectl delete crd desires.manifestation.pez.sh`.

### Install with kustomize

```sh
# Build and push your image
make docker-build docker-push IMG=<your-registry>/manifestation-operator:latest

# Install the CRD and deploy the operator (webhook + cert-manager wired in)
make deploy IMG=<your-registry>/manifestation-operator:latest
```

### Run the demo

```sh
kubectl apply -f examples/demo.yaml
kubectl -n manifest-demo get pods      # nginx is Pending / SchedulingGated
kubectl -n manifest-demo get desire    # watch it flip to Manifested
kubectl -n manifest-demo get pods      # nginx is Running
```

By default every namespace is subject to manifestation except the control-plane ones
(`kube-system`, `kube-node-lease`, `kube-public`, `cert-manager`). A single pod can
refuse the journey with the label `manifestation.pez.sh/skip=true`, if it is the kind of
pod that does not believe in anything.

Prefer the cautious life? Switch back to opt-in by setting the webhook to match only
labelled namespaces:

```sh
# Helm
helm upgrade ... --set-json 'webhook.namespaceSelector={"matchLabels":{"manifestation.pez.sh/enabled":"true"}}'
# then label the namespaces you want gated
kubectl label namespace <ns> manifestation.pez.sh/enabled=true
```

### Local development

```sh
make manifests generate   # regenerate CRD, RBAC, webhook config, deepcopy
make test                 # unit + envtest
go test ./internal/manifest/...   # just the affirmation validator
make run                  # run the controller locally (webhook needs certs)
```

## Safety

The webhook's `failurePolicy` is `Ignore`, so if the operator is down pods schedule
normally (fail-open) instead of pod creation freezing. Control-plane namespaces are
always excluded and the operator self-skips its own pod via
`manifestation.pez.sh/skip=true` - otherwise it could never start to release anyone.
Applying to the whole cluster is the fun setting; it is also the one that turns a quiet
afternoon into an incident, so know your cluster before you flip it on in production.

## License

Apache 2.0. Manifest responsibly.
