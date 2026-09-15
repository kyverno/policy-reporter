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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type EmailReportType string

const (
	EmailReportTypeSummary    EmailReportType = "summary"
	EmailReportTypeViolations EmailReportType = "violations"
)

// EmailReportSpec defines a scheduled email report for one namespace.
type EmailReportSpec struct {
	// Type selects the existing summary or violations email report.
	// +kubebuilder:validation:Enum=summary;violations
	Type EmailReportType `json:"type"`

	// Schedule is a cron expression accepted by Kubernetes CronJobs.
	// +kubebuilder:validation:MinLength=1
	Schedule string `json:"schedule"`

	// To lists the report recipients.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:MinLength=1
	To []string `json:"to"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=emailreports,scope=Namespaced,shortName=er
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// EmailReport configures a scheduled email report scoped to its namespace.
type EmailReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EmailReportSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EmailReportList contains a list of EmailReport resources.
type EmailReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EmailReport `json:"items"`
}
