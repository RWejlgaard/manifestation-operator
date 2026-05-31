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

1. A **mutating admission webhook** intercepts every pod created in a namespace
   labelled `manifestation.pez.sh/enabled=true` and injects a
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

```sh
# Build and push your image
make docker-build docker-push IMG=<your-registry>/manifestation-operator:latest

# Install the CRD and deploy the operator (webhook + cert-manager wired in)
make deploy IMG=<your-registry>/manifestation-operator:latest

# Run the demo
kubectl apply -f examples/demo.yaml
kubectl -n manifest-demo get pods      # nginx is Pending / SchedulingGated
kubectl -n manifest-demo get desire    # watch it flip to Manifested
kubectl -n manifest-demo get pods      # nginx is Running
```

To opt a namespace into manifestation:

```sh
kubectl label namespace <ns> manifestation.pez.sh/enabled=true
```

A single pod can refuse the journey with the label `manifestation.pez.sh/skip=true`,
if it is the kind of pod that does not believe in anything.

### Local development

```sh
make manifests generate   # regenerate CRD, RBAC, webhook config, deepcopy
make test                 # unit + envtest
go test ./internal/manifest/...   # just the affirmation validator
make run                  # run the controller locally (webhook needs certs)
```

## Safety

The webhook's `failurePolicy` is `Ignore` and it is scoped by `namespaceSelector` to
opted-in namespaces only. If the operator is down, pod creation everywhere else is
unaffected - manifestation is opt-in, not a cluster-wide hostage situation. Funny once,
an incident twice.

## License

Apache 2.0. Manifest responsibly.
