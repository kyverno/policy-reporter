/*
Copyright 2020 The Kubernetes authors.

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

package v1alpha2

import (
	"strconv"

	"github.com/segmentio/fasthash/fnv1a"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	reportsv1alpha1 "openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:path=clusterpolicyreports,scope="Cluster",shortName=cpolr
// +kubebuilder:printcolumn:name="Kind",type=string,JSONPath=`.scope.kind`,priority=1
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.scope.name`,priority=1
// +kubebuilder:printcolumn:name="Pass",type=integer,JSONPath=`.summary.pass`
// +kubebuilder:printcolumn:name="Fail",type=integer,JSONPath=`.summary.fail`
// +kubebuilder:printcolumn:name="Warn",type=integer,JSONPath=`.summary.warn`
// +kubebuilder:printcolumn:name="Error",type=integer,JSONPath=`.summary.error`
// +kubebuilder:printcolumn:name="Skip",type=integer,JSONPath=`.summary.skip`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClusterPolicyReport is the Schema for the clusterpolicyreports API
type ClusterPolicyReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Scope is an optional reference to the report scope (e.g. a Deployment, Namespace, or Node)
	// +optional
	Scope *corev1.ObjectReference `json:"scope,omitempty"`

	// ScopeSelector is an optional selector for multiple scopes (e.g. Pods).
	// Either one of, or none of, but not both of, Scope or ScopeSelector should be specified.
	// +optional
	ScopeSelector *metav1.LabelSelector `json:"scopeSelector,omitempty"`

	// PolicyReportSummary provides a summary of results
	// +optional
	Summary PolicyReportSummary `json:"summary,omitempty"`

	// PolicyReportResult provides result details
	// +optional
	Results []PolicyReportResult `json:"results,omitempty"`
}

func (r *ClusterPolicyReport) GetResults() []PolicyReportResult {
	return r.Results
}

func (r *ClusterPolicyReport) HasResult(id string) bool {
	for _, r := range r.Results {
		if r.GetID() == id {
			return true
		}
	}

	return false
}

func (r *ClusterPolicyReport) SetResults(results []PolicyReportResult) {
	r.Results = results
}

func (r *ClusterPolicyReport) GetSummary() PolicyReportSummary {
	return r.Summary
}

func (r *ClusterPolicyReport) GetSource() string {
	if len(r.Results) == 0 {
		return ""
	}

	return r.Results[0].Source
}

func (r *ClusterPolicyReport) GetID() string {
	h1 := fnv1a.Init64
	h1 = fnv1a.AddString64(h1, r.GetName())

	return strconv.FormatUint(h1, 10)
}

func (r *ClusterPolicyReport) GetKey() string {
	return r.Name
}

func (r *ClusterPolicyReport) GetKinds() []string {
	if r.GetScope() != nil {
		return []string{r.Scope.Kind}
	}

	list := make([]string, 0)
	for _, k := range r.Results {
		if !k.HasResource() {
			continue
		}

		kind := k.GetResource().Kind

		if kind == "" || helper.Contains(kind, list) {
			continue
		}

		list = append(list, kind)
	}

	return list
}

func (r *ClusterPolicyReport) GetSeverities() []string {
	list := make([]string, 0)
	for _, k := range r.Results {

		if k.Severity == "" || helper.Contains(string(k.Severity), list) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func (r *ClusterPolicyReport) GetScope() *corev1.ObjectReference {
	return r.Scope
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterPolicyReportList contains a list of ClusterPolicyReport
type ClusterPolicyReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPolicyReport `json:"items"`
}

func (polr *ClusterPolicyReport) ToOpenReports() *reportsv1alpha1.ClusterReport {
	res := []reportsv1alpha1.ReportResult{}
	for _, r := range polr.GetResults() {
		res = append(res, reportsv1alpha1.ReportResult{
			Source:           r.Source,
			Policy:           r.Policy,
			Rule:             r.Rule,
			Category:         r.Category,
			Timestamp:        r.Timestamp,
			Severity:         reportsv1alpha1.ResultSeverity(r.Severity),
			Result:           reportsv1alpha1.Result(r.Result),
			Subjects:         r.Resources,
			ResourceSelector: r.ResourceSelector,
			Scored:           r.Scored,
			Description:      r.Message,
			Properties:       r.Properties,
		})
	}
	return &reportsv1alpha1.ClusterReport{
		ObjectMeta: v1.ObjectMeta{
			Name:      polr.Name,
			Namespace: polr.Namespace,
		},
		Source:        polr.GetSource(),
		Scope:         polr.Scope,
		ScopeSelector: polr.ScopeSelector,
		Summary: reportsv1alpha1.ReportSummary{
			Pass:  polr.Summary.Pass,
			Fail:  polr.Summary.Fail,
			Warn:  polr.Summary.Warn,
			Error: polr.Summary.Error,
			Skip:  polr.Summary.Skip,
		},
		Results: res,
	}
}
