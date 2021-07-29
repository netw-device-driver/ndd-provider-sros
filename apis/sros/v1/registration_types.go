/*
Copyright 2021 Wim Henderickx.

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
	nddv1 "github.com/netw-device-driver/ndd-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DeviceMatch                     = "sros_nokia"
	DeviceTypeSROS nddv1.DeviceType = "nokia-sros"
)

type RegistrationParameters struct {
	// Registrations defines the Registrations the device driver subscribes to for config change notifications
	// +optional
	Subscriptions []string `json:"subscriptions,omitempty"`

	// ExceptionPaths defines the exception paths that should be ignored during change notifications
	// if the xpath contains the exception path it is considered a match
	// +optional
	ExceptionPaths []string `json:"exceptionPaths,omitempty"`

	// ExplicitExceptionPaths defines the exception paths that should be ignored during change notifications
	// the match should be exact to condider this xpath
	// +optional
	ExplicitExceptionPaths []string `json:"explicitExceptionPaths,omitempty"`
}

// RegistrationObservation are the observable fields of a Registration.
type RegistrationObservation struct {
	//ObservableField string `json:"observableField,omitempty"`
}

// A RegistrationSpec defines the desired state of a Registration.
type RegistrationSpec struct {
	nddv1.ResourceSpec `json:",inline"`
	ForProvider        RegistrationParameters `json:"forProvider"`
}

// A RegistrationStatus represents the observed state of a Registration.
type RegistrationStatus struct {
	nddv1.ResourceStatus `json:",inline"`
	AtProvider           RegistrationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Registration is the Schema for the Registrations API
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TARGET",type="string",JSONPath=".status.conditions[?(@.kind=='TargetFound')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={sros},shortName=srosreg
type Registration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrationSpec   `json:"spec"`
	Status RegistrationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RegistrationList contains a list of Registration
type RegistrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Registration{}, &RegistrationList{})
}

func (o *Registration) GetSubscriptions() []string {
	return o.Spec.ForProvider.Subscriptions
}

func (o *Registration) SetSubscriotions(sub []string) {
	o.Spec.ForProvider.Subscriptions = sub
}

func (o *Registration) GetExceptionPaths() []string {
	return o.Spec.ForProvider.ExceptionPaths
}

func (o *Registration) SetExceptionPaths(ep []string) {
	o.Spec.ForProvider.ExceptionPaths = ep
}

func (o *Registration) GetExplicitExceptionPaths() []string {
	return o.Spec.ForProvider.ExceptionPaths
}

func (o *Registration) SetExplicitExceptionPaths(eep []string) {
	o.Spec.ForProvider.ExceptionPaths = eep
}
