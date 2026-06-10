/*
Copyright 2025.

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

// +kubebuilder:oneOf:={required:{s3}}
// +kubebuilder:oneOf:={required:{webhook}}
// +kubebuilder:oneOf:={required:{telegram}}
// +kubebuilder:oneOf:={required:{slack}}
// +kubebuilder:oneOf:={required:{elasticSearch}}
// +kubebuilder:oneOf:={required:{gcs}}
// +kubebuilder:oneOf:={required:{loki}}
// +kubebuilder:oneOf:={required:{securityHub}}
// +kubebuilder:oneOf:={required:{kinesis}}
// +kubebuilder:oneOf:={required:{splunk}}
// +kubebuilder:oneOf:={required:{teams}}
// +kubebuilder:oneOf:={required:{jira}}
// +kubebuilder:oneOf:={required:{alertManager}}

// TargetConfigSpec defines the desired state of TargetConfig.
type TargetConfigSpec struct {
	Config `json:",inline"`

	//+optional
	S3 *S3Options `json:"s3,omitempty"`

	// +optional
	Webhook *WebhookOptions `json:"webhook,omitempty"`

	// +optional
	Telegram *TelegramOptions `json:"telegram,omitempty"`

	// +optional
	Slack *SlackOptions `json:"slack,omitempty"`

	// +optional
	ElasticSearch *ElasticsearchOptions `json:"elasticSearch,omitempty"`

	// +optional
	GCS *GCSOptions `json:"gcs,omitempty"`

	// +optional
	Loki *LokiOptions `json:"loki,omitempty"`

	// +optional
	SecurityHub *SecurityHubOptions `json:"securityHub,omitempty"`

	// +optional
	Kinesis *KinesisOptions `json:"kinesis,omitempty"`

	// +optional
	Teams *WebhookOptions `json:"teams,omitempty"`

	// +optional
	Jira *JiraOptions `json:"jira,omitempty"`

	// +optional
	AlertManager *HostOptions `json:"alertManager,omitempty"`

	// +optional
	Splunk *SplunkOptions `json:"splunk,omitempty"`

	// +kubebuilder:default=true
	// +optional
	SkipExisting bool `json:"skipExistingOnStartup,omitempty"`
}

// TargetConfigStatus defines the observed state of TargetConfig.
type TargetConfigStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=targetconfigs,scope=Cluster,shortName=tcfg
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient

// TargetConfig is the Schema for the targetconfigs API.
type TargetConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TargetConfigSpec   `json:"spec,omitempty"`
	Status TargetConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TargetConfigList contains a list of TargetConfig.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TargetConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TargetConfig `json:"items"`
}
