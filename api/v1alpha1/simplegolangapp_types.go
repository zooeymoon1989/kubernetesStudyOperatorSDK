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

// SimpleGolangAppSpec defines the desired state of SimpleGolangApp.
type SimpleGolangAppSpec struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Optional
	// Image is the container image to run
	Image string `json:"image,omitempty"`

	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	// Replicas is the desired number of pods
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:default:=80
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Optional
	// Port is the container port and service port
	Port *int32 `json:"port,omitempty"`
}

// SimpleGolangAppStatus defines the observed state of SimpleGolangApp.
type SimpleGolangAppStatus struct {
	// +kubebuilder:validation:Optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// +kubebuilder:validation:Optional
	ServiceName string `json:"serviceName,omitempty"`
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations of an object's state.
	// +kubebuilder:validation:Optional
	// +kubebuilder:patchStrategy=merge
	// +kubebuilder:patchMergeKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.status.serviceName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SimpleGolangApp is the Schema for the simplegolangapps API.
type SimpleGolangApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleGolangAppSpec   `json:"spec,omitempty"`
	Status SimpleGolangAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SimpleGolangAppList contains a list of SimpleGolangApp.
type SimpleGolangAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleGolangApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleGolangApp{}, &SimpleGolangAppList{})
}
