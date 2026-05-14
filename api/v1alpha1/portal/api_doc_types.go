package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// ApiDocConditionReady is the condition type for ApiDoc readiness.
	ApiDocConditionReady = "Ready"
)

// Reason constants for ApiDoc conditions.
const (
	// ApiDocReasonReady is the condition reason for when the ApiDoc is ready.
	ApiDocReasonReady = "Ready"
	// ApiDocReasonInvalid is the condition reason for when the ApiDoc is invalid. This
	// can be set for any reason that prevents the ApiDoc from being accepted including
	// fetch failures, schema validation errors, size limits, servedBy target not found, etc.
	// The specific error details are provided in the condition Message field.
	ApiDocReasonInvalid = "Invalid"
)

// MaxStitchedContentLength is the maximum size in bytes of StitchedSource.Content.
// MUST stay in sync with the +kubebuilder:validation:MaxLength marker on
// StitchedSource.Content below; the marker is parsed as literal text by codegen
// and cannot reference this constant.
const MaxStitchedContentLength = 700000

// +kubebuilder:rbac:groups=portal.solo.io,resources=apidocs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=portal.solo.io,resources=apidocs/status,verbs=get;update;patch

// ApiDoc represents an API document containing an OpenAPI schema.
// The schema can be provided inline, fetched from a URL, or fetched from
// a supported in-cluster resource (e.g., Service or Backend).
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={portal},path=apidocs
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=".status.sourceType",description="Source type (Kube, Url, Manual, Stitched)"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status",description="ApiDoc readiness status"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp"
type ApiDoc struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of the ApiDoc.
	// +required
	Spec ApiDocSpec `json:"spec"`

	// status is the current state of the ApiDoc.
	// +optional
	Status ApiDocStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// ApiDocSpec defines the desired state of ApiDoc.
// +kubebuilder:validation:XValidation:rule="!(has(self.source.url) || has(self.source.manual)) || has(self.servedBy)",message="servedBy is required for manual and url sources"
type ApiDocSpec struct {
	// source defines where to obtain the OpenAPI schema.
	// +required
	Source ApiDocSource `json:"source"`

	// servedBy anchors this API doc to a backend for Try-It-Out functionality.
	// Required for manual and url sources to enable routing resolution.
	// For kube sources, if omitted, the controller will infer servedBy from
	// spec.source.kube.targetRef. In case you want to override the inferred servedBy,
	// you can set servedBy explicitly. It supports any GVK that contributes to the kgateway
	// routing model, including core Services and kgateway Backends.
	// +optional
	ServedBy *ServedBy `json:"servedBy,omitempty"`

	// retry overrides the controller's schema fetch retry behavior for url and
	// kube sources. Any field left unset inherits the controller-level default.
	// +optional
	Retry *ApiDocRetryStrategy `json:"retry,omitempty"`
}

// ApiDocRetryStrategy configures retries for URL and kube-backed schema fetches.
type ApiDocRetryStrategy struct {
	// maxAttempts is the maximum number of fetch attempts.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxAttempts *int32 `json:"maxAttempts,omitempty"`

	// delay is the delay between retry attempts.
	// +optional
	// +kubebuilder:validation:XValidation:rule="matches(self, '^([0-9]{1,5}(h|m|s|ms)){1,4}$')",message="invalid duration value"
	// +kubebuilder:validation:XValidation:rule="duration(self) >= duration('0s')",message="delay must be non-negative."
	Delay *metav1.Duration `json:"delay,omitempty"`

	// useBackoff controls whether retries use exponential backoff.
	// +optional
	UseBackoff *bool `json:"useBackoff,omitempty"`
}

// ApiDocSource defines the source of the OpenAPI schema.
// Exactly one source type must be specified.
// +kubebuilder:validation:ExactlyOneOf=manual;url;kube;stitched
type ApiDocSource struct {
	// manual specifies the OpenAPI schema content directly in the spec.
	// Use for small, static schemas that don't change frequently.
	// +optional
	Manual *ManualSource `json:"manual,omitempty"`

	// url specifies an internal or external URL to fetch the OpenAPI schema from.
	// +optional
	URL *URLSource `json:"url,omitempty"`

	// kube specifies a Kubernetes target (Service or Backend) to fetch
	// the OpenAPI schema from via HTTP.
	// +optional
	Kube *KubeSource `json:"kube,omitempty"`

	// stitched specifies that this ApiDoc contains a stitched/projected schema
	// generated by the ApiProduct controller. The schema is derived from one or
	// more source ApiDocs and has paths rewritten according to the ApiProduct's
	// routing configuration. This source type is controller-managed and should
	// not be set by users directly.
	// +optional
	Stitched *StitchedSource `json:"stitched,omitempty"`
}

// ManualSource specifies the OpenAPI schema content directly.
type ManualSource struct {
	// content is the raw OpenAPI schema in JSON or YAML format.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=700000
	Content string `json:"content"`
}

// URLSource specifies an external URL to fetch the OpenAPI schema from.
type URLSource struct {
	// endpoint is the URL to fetch the OpenAPI schema from.
	// Must be a valid HTTP or HTTPS URL.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^https?://`
	Endpoint string `json:"endpoint"`
}

// KubeSource specifies an in-cluster Kubernetes target to fetch OpenAPI from.
type KubeSource struct {
	// targetRef references the Kubernetes resource to fetch the OpenAPI schema from.
	// Currently supports core Services or gateway.kgateway.dev Backends.
	// +required
	// +kubebuilder:validation:XValidation:rule="(self.group == '' && self.kind == 'Service') || (self.group == 'gateway.kgateway.dev' && self.kind == 'Backend')",message="targetRef must reference a core Service or a gateway.kgateway.dev Backend."
	TargetRef gwv1.BackendObjectReference `json:"targetRef"`
	// path is the HTTP path to fetch the OpenAPI schema from.
	// Example: "/openapi", "/swagger.json", "/v3/api-docs"
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:Pattern=`^/`
	Path string `json:"path"`
}

// StitchedSource indicates this ApiDoc contains a controller-generated stitched schema.
// The schema is projected from source ApiDocs according to ApiProduct routing configuration.
type StitchedSource struct {
	// content is the stitched OpenAPI schema in JSON format.
	// This is generated by the ApiProduct controller and should not be modified directly.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=700000
	// MaxLength MUST stay in sync with MaxStitchedContentLength.
	Content string `json:"content"`
}

// ServedBy specifies the backend that this ApiDoc is anchored to.
type ServedBy struct {
	// +kubebuilder:validation:XValidation:rule="(self.group == '' && self.kind == 'Service') || (self.group == 'gateway.kgateway.dev' && self.kind == 'Backend') || (self.group == 'portal.solo.io' && self.kind == 'Stitched')",message="servedBy must reference a core Service, gateway.kgateway.dev Backend, or portal.solo.io Stitched."
	gwv1.BackendObjectReference `json:",inline"`
}

// ApiDocStatus defines the observed state of ApiDoc.
type ApiDocStatus struct {
	// conditions represent the current state of the ApiDoc.
	// The "Ready" condition indicates whether the schema is valid and available.
	// +listType=map
	// +listMapKey=type
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// sourceType indicates the source type of this ApiDoc (kube, url, manual, stitched).
	// This is set by the controller based on spec.source and used for display purposes.
	// +optional
	SourceType *string `json:"sourceType,omitempty"`

	// resolvedSchema contains the resolved OpenAPI schema. This is the raw schema
	// that has been normalized to OpenAPI 3.x JSON.
	// +optional
	ResolvedSchema *string `json:"resolvedSchema,omitempty"`
}

// ApiDocList contains a list of ApiDoc resources.
// +kubebuilder:object:root=true
type ApiDocList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiDoc `json:"items"`
}
