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

package v1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/rwejlgaard/manifestation-operator/internal/manifest"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &corev1.Pod{}).
		WithDefaulter(&PodCustomDefaulter{}).
		Complete()
}

// The namespaceSelector on the generated webhook config (config/webhook/manifests.yaml)
// limits this to namespaces labelled manifestation.pez.sh/enabled=true, so an outage of
// the webhook cannot freeze pod creation cluster-wide.
// +kubebuilder:webhook:path=/mutate-core-k8s-io-v1-pod,mutating=true,failurePolicy=ignore,sideEffects=None,groups="",resources=pods,verbs=create,versions=v1,name=mpod-v1.kb.io,admissionReviewVersions=v1

// PodCustomDefaulter holds a pod in limbo by injecting a scheduling gate at creation
// time. The pod stays Pending until a worthy Desire removes the gate.
type PodCustomDefaulter struct{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for Pod.
func (d *PodCustomDefaulter) Default(_ context.Context, pod *corev1.Pod) error {
	// A pod can explicitly opt out. Free will exists, even here.
	if pod.Labels[manifest.SkipLabel] == "true" {
		return nil
	}

	// Already gated? The universe does not gate twice.
	for _, g := range pod.Spec.SchedulingGates {
		if g.Name == manifest.SchedulingGate {
			return nil
		}
	}

	podlog.Info("placing pod in manifestation limbo", "name", pod.GetName(), "namespace", pod.GetNamespace())
	pod.Spec.SchedulingGates = append(pod.Spec.SchedulingGates, corev1.PodSchedulingGate{
		Name: manifest.SchedulingGate,
	})
	return nil
}
