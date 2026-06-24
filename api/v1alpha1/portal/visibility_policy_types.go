package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups=portal.solo.io,resources=visibilitypolicies,verbs=get;list;watch

// VisibilityPolicy defines a set of claim-group requirements that may be
// referenced from a Portal to gate access to ApiProducts.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={portal},path=visibilitypolicies,shortName=vp
// +kubebuilder:metadata:labels={app=gloo,app.kubernetes.io/name=visibilitypolicies}
type VisibilityPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of the VisibilityPolicy
	// +required
	Spec VisibilityPolicySpec `json:"spec"`
}

// VisibilityPolicySpec defines the claim-group requirements granted by this policy.
type VisibilityPolicySpec struct {
	// membership lists the claim-group requirements granted by this policy.
	// OR across groups; AND within a group.
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=64
	// +required
	Membership []ClaimGroup `json:"membership"`
}

// VisibilityPolicyList contains a list of VisibilityPolicy resources
// +kubebuilder:object:root=true
type VisibilityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VisibilityPolicy `json:"items"`
}
