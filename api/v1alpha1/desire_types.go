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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Intensity is how hard you believe. Purely vibes, but the universe is listening.
// +kubebuilder:validation:Enum=whisper;speak;shout;chant
type Intensity string

const (
	IntensityWhisper Intensity = "whisper"
	IntensitySpeak   Intensity = "speak"
	IntensityShout   Intensity = "shout"
	IntensityChant   Intensity = "chant"
)

// DesirePhase is where your manifestation stands with the universe.
// +kubebuilder:validation:Enum=Pending;Manifested;Rejected
type DesirePhase string

const (
	// PhasePending means the universe has heard you but pods are not yet manifested.
	PhasePending DesirePhase = "Pending"
	// PhaseManifested means the affirmation was accepted and matching pods are released.
	PhaseManifested DesirePhase = "Manifested"
	// PhaseRejected means the affirmation was not in present tense. The universe rejects doubt.
	PhaseRejected DesirePhase = "Rejected"
)

// DesireSpec defines the desired state of Desire.
type DesireSpec struct {
	// manifestation is your affirmation. It MUST be written in present tense, as if it
	// is already true. "My nginx pod is healthy and serving traffic." Not "will be".
	// The universe does not respond to doubt, future tense, or conditionals.
	// +required
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=512
	Manifestation string `json:"manifestation"`

	// selector picks which pods in this namespace your desire manifests. If empty, your
	// desire radiates to every awaiting pod in the namespace. Choose your energy.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// intensity is how hard you believe. It does not change physics. It changes you.
	// +optional
	// +kubebuilder:default=speak
	Intensity Intensity `json:"intensity,omitempty"`
}

// DesireStatus defines the observed state of Desire.
type DesireStatus struct {
	// phase is where this desire stands with the universe.
	// +optional
	Phase DesirePhase `json:"phase,omitempty"`

	// reason is what the universe whispers back.
	// +optional
	Reason string `json:"reason,omitempty"`

	// manifestedPods is how many pods this desire has released from limbo.
	// +optional
	ManifestedPods int32 `json:"manifestedPods,omitempty"`

	// observedGeneration is the generation of the Desire spec last reconciled.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// conditions represent the current state of the Desire resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Pods",type=integer,JSONPath=`.status.manifestedPods`
// +kubebuilder:printcolumn:name="Manifestation",type=string,JSONPath=`.spec.manifestation`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Desire is the Schema for the desires API.
type Desire struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Desire
	// +required
	Spec DesireSpec `json:"spec"`

	// status defines the observed state of Desire
	// +optional
	Status DesireStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DesireList contains a list of Desire.
type DesireList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Desire `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Desire{}, &DesireList{})
}
