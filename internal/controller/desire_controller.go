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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	manifestationv1alpha1 "github.com/rwejlgaard/manifestation-operator/api/v1alpha1"
	"github.com/rwejlgaard/manifestation-operator/internal/manifest"
)

// resyncInterval re-checks a manifested desire so pods created after it (and freshly
// gated by the webhook) still get released even if the pod watch misses them.
const resyncInterval = 30 * time.Second

// DesireReconciler reconciles a Desire object.
type DesireReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=manifestation.pez.sh,resources=desires,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manifestation.pez.sh,resources=desires/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=manifestation.pez.sh,resources=desires/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile validates a Desire's affirmation and, if the universe accepts it, releases
// the matching gated pods from limbo.
func (r *DesireReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var desire manifestationv1alpha1.Desire
	if err := r.Get(ctx, req.NamespacedName, &desire); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Judge the affirmation.
	ok, reason := manifest.Validate(desire.Spec.Manifestation)
	if !ok {
		log.Info("desire rejected", "reason", reason)
		return r.finish(ctx, &desire, manifestationv1alpha1.PhaseRejected, reason, 0)
	}

	// Find the pods this desire speaks to.
	selector := labels.Everything()
	if desire.Spec.Selector != nil {
		s, err := metav1.LabelSelectorAsSelector(desire.Spec.Selector)
		if err != nil {
			return r.finish(ctx, &desire, manifestationv1alpha1.PhaseRejected,
				"Your selector is malformed. Even the universe needs valid labels.", 0)
		}
		selector = s
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods,
		client.InNamespace(desire.Namespace),
		client.MatchingLabelsSelector{Selector: selector},
	); err != nil {
		return ctrl.Result{}, err
	}

	released := 0
	for i := range pods.Items {
		pod := &pods.Items[i]
		if !hasGate(pod) {
			continue
		}
		patch := client.MergeFrom(pod.DeepCopy())
		pod.Spec.SchedulingGates = withoutGate(pod.Spec.SchedulingGates)
		if pod.Annotations == nil {
			pod.Annotations = map[string]string{}
		}
		pod.Annotations[manifest.ManifestedByAnnotation] = desire.Name
		if err := r.Patch(ctx, pod, patch); err != nil {
			if apierrors.IsConflict(err) || apierrors.IsNotFound(err) {
				continue // someone raced us; reconcile again later
			}
			log.Error(err, "failed to manifest pod", "pod", pod.Name)
			return ctrl.Result{}, err
		}
		log.Info("manifested pod", "pod", pod.Name)
		released++
	}

	result, err := r.finish(ctx, &desire, manifestationv1alpha1.PhaseManifested, reason,
		desire.Status.ManifestedPods+int32(released))
	if err != nil {
		return result, err
	}
	// Keep watching for future pods that need the same blessing.
	return ctrl.Result{RequeueAfter: resyncInterval}, nil
}

// finish writes the desire's status (phase, reason, condition, counters) and returns.
func (r *DesireReconciler) finish(ctx context.Context, desire *manifestationv1alpha1.Desire,
	phase manifestationv1alpha1.DesirePhase, reason string, manifested int32) (ctrl.Result, error) {

	desire.Status.Phase = phase
	desire.Status.Reason = reason
	desire.Status.ManifestedPods = manifested
	desire.Status.ObservedGeneration = desire.Generation

	condStatus := metav1.ConditionTrue
	condReason := "Manifested"
	if phase != manifestationv1alpha1.PhaseManifested {
		condStatus = metav1.ConditionFalse
		condReason = "NotPresentTense"
	}
	meta.SetStatusCondition(&desire.Status.Conditions, metav1.Condition{
		Type:    "Manifested",
		Status:  condStatus,
		Reason:  condReason,
		Message: reason,
	})

	if err := r.Status().Update(ctx, desire); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, nil
}

func hasGate(pod *corev1.Pod) bool {
	for _, g := range pod.Spec.SchedulingGates {
		if g.Name == manifest.SchedulingGate {
			return true
		}
	}
	return false
}

func withoutGate(gates []corev1.PodSchedulingGate) []corev1.PodSchedulingGate {
	out := gates[:0]
	for _, g := range gates {
		if g.Name != manifest.SchedulingGate {
			out = append(out, g)
		}
	}
	return out
}

// podToDesires enqueues every Desire in a pod's namespace whenever that pod changes, so
// a newly gated pod is promptly judged by any standing affirmations.
func (r *DesireReconciler) podToDesires(ctx context.Context, obj client.Object) []reconcile.Request {
	var desires manifestationv1alpha1.DesireList
	if err := r.List(ctx, &desires, client.InNamespace(obj.GetNamespace())); err != nil {
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(desires.Items))
	for i := range desires.Items {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: client.ObjectKeyFromObject(&desires.Items[i]),
		})
	}
	return reqs
}

// SetupWithManager sets up the controller with the Manager.
func (r *DesireReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&manifestationv1alpha1.Desire{}).
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.podToDesires)).
		Named("desire").
		Complete(r)
}
