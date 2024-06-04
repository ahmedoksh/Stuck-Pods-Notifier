/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StuckPodsNotifierSpec defines the desired state of StuckPodsNotifier
type StuckPodsNotifierSpec struct {
	//Label selector to select pods to monitor
	//+kubebuilder:validation:Required
	Selector metav1.LabelSelector `json:"selector"`

	//Slack token to use for sending notifications
	//+kubebuilder:validation:Required
	SlackToken string `json:"slackToken"`

	//Slack channel to send notifications to
	//+kubebuilder:validation:Required
	SlackChannel string `json:"slackChannel"`

	//How often to check for stuck pods
	//+kubebuilder:validation:Format:=duration
	//+kubebuilder:default:="1m"
	RequeueAfter string `json:"requeueAfter,omitempty"`

	//The time after which a pod is considered stuck
	//+kubebuilder:validation:Format:=duration
	//+kubebuilder:default:="2m"
	PodWaitThreshold string `json:"podWaitThreshold,omitempty"`
}

type StuckPodDetail struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	CreationTime string `json:"creationTime"`
}

// StuckPodsNotifierStatus defines the observed state of StuckPodsNotifier
type StuckPodsNotifierStatus struct {
	PendingPodsCount int              `json:"pendingPodsCount"`
	StuckPodsCount   int              `json:"stuckPodsCount"`
	StuckPodsDetails []StuckPodDetail `json:"stuckPodsDetails"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StuckPodsNotifier is the Schema for the stuckpodsnotifiers API
type StuckPodsNotifier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StuckPodsNotifierSpec   `json:"spec,omitempty"`
	Status StuckPodsNotifierStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StuckPodsNotifierList contains a list of StuckPodsNotifier
type StuckPodsNotifierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StuckPodsNotifier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StuckPodsNotifier{}, &StuckPodsNotifierList{})
}
