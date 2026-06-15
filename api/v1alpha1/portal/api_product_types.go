package v1alpha1

import (
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:rbac:groups=portal.solo.io,resources=apiproducts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=portal.solo.io,resources=apiproducts/status,verbs=get;update;patch

// +kubebuilder:printcolumn:name="ID",type=string,JSONPath=".spec.id",description="API Product ID"
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=".spec.displayName",description="Display name"

// ApiProduct groups one or more API versions together as a logical product
// for exposure in the developer portal.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={portal},path=apiproducts
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels={app.kubernetes.io/name=apiproduct}
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status",description="ApiProduct readiness status"
type ApiProduct struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of the ApiProduct
	// +required
	Spec ApiProductSpec `json:"spec"`

	// status defines the observed state of the ApiProduct
	// +optional
	Status ApiProductStatus `json:"status,omitempty"` // nolint:kubeapilinter // optionalfields - allow status to be a non-pointer
}

// ApiProductSpec defines the desired state of ApiProduct.
type ApiProductSpec struct {
	// id is a URL-safe, unique identifier for this API product in the portal.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`
	// +required
	ID string `json:"id"`

	// displayName is the name for the API product to display in the frontend portal.
	// If omitted, the ID value is used as the display name.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=128
	// +optional
	DisplayName *string `json:"displayName,omitempty"`

	// versions is the list of versions of the API product.
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:validation:XValidation:rule="self.all(v1, self.exists_one(v2, v1.name == v2.name))",message="version names must be unique"
	// +required
	Versions []Version `json:"versions"`

	// customMetadata provides key-value pairs of custom metadata to display in the developer portal.
	// This metadata applies to the API product as a whole (i.e., across all versions).
	// +kubebuilder:validation:MinProperties=1
	// +kubebuilder:validation:MaxProperties=16
	// +optional
	CustomMetadata map[string]string `json:"customMetadata,omitempty"`
}

// Version defines a version of an API product.
type Version struct {
	// name is the version identifier (e.g., "v1", "v2", "1.0.0").
	// It must be unique within the ApiProduct's versions list.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`
	// +required
	Name string `json:"name"`

	// targetRefs references the HTTPRoutes that expose this API version.
	// Only HTTPRoutes attached to kgateway enterprise Gateways are currently supported.
	//
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:validation:XValidation:rule="self.all(r, r.group == 'gateway.networking.k8s.io' && r.kind == 'HTTPRoute')",message="targetRefs must reference HTTPRoute resources from gateway.networking.k8s.io"
	// +kubebuilder:validation:XValidation:rule="self.all(r1, self.filter(r2, r1.name == r2.name).size() == 1 || (r1.apiDocSelectors.size() > 0 && self.filter(r2, r2.name == r1.name && r2.apiDocSelectors == r1.apiDocSelectors).size() == 1))",message="duplicate targetRef names require each entry to specify distinct non-empty apiDocSelectors"
	// +required
	TargetRefs []PortalTargetReference `json:"targetRefs"` // nolint:kubeapilinter // arrayofstruct - required fields inherited from embedded LocalPolicyTargetReference

	// openApiMetadata sets the OpenAPI specification metadata for the API product version.
	// +optional
	OpenAPIMetadata *OpenAPIMetadata `json:"openApiMetadata,omitempty"`

	// customMetadata provides key-value pairs of custom metadata to display in the developer portal.
	// This metadata applies to this specific version of the API product.
	// +kubebuilder:validation:MinProperties=1
	// +kubebuilder:validation:MaxProperties=16
	// +optional
	CustomMetadata map[string]string `json:"customMetadata,omitempty"`
}

// PortalTargetReference extends LocalPolicyTargetReference with portal-specific fields
// for controlling ApiDoc selection when multiple ApiDocs serve the same backend.
type PortalTargetReference struct {
	upstreamshared.LocalPolicyTargetReference `json:",inline"`

	// apiDocSelectors specifies which ApiDocs to use for schema stitching when
	// multiple ApiDocs serve the same backend referenced by this HTTPRoute.
	// Each selector references an ApiDoc by name and namespace; the target backend
	// is inferred from the ApiDoc's spec.servedBy field.
	//
	// When omitted, automatic 1:1 backend-to-ApiDoc resolution is used.
	// If multiple ApiDocs match the same backend and no selector is provided,
	// the version fails with an error.
	//
	// +listType=map
	// +listMapKey=name
	// +listMapKey=namespace
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +optional
	ApiDocSelectors []ApiDocSelector `json:"apiDocSelectors,omitempty"`
}

// ApiDocSelector identifies an ApiDoc by name and namespace for use in apiDocSelectors.
type ApiDocSelector struct {
	// name is the name of the ApiDoc to select.
	// +kubebuilder:validation:MinLength=1
	// +required
	Name string `json:"name"`

	// namespace is the namespace of the ApiDoc to select.
	// +kubebuilder:validation:MinLength=1
	// +required
	Namespace string `json:"namespace"`
}

// OpenAPIMetadata is the metadata for the OpenAPI specification for a given API product version.
// When configured, at least one field should be set.
// All fields except contact are surfaced as top-level fields on the API product version
// response (description as "documentation"). All fields are also merged into the info object
// of the stitched OpenAPI schema returned in apiSpec: title, description, and termsOfService
// map to info.title, info.description, and info.termsOfService; contact maps to
// info.contact.url; and license maps to info.license.name.
type OpenAPIMetadata struct {
	// title is the title of the OpenAPI specification for this API.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +optional
	Title *string `json:"title,omitempty"`

	// description is the description of the OpenAPI specification for this API.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +optional
	Description *string `json:"description,omitempty"`

	// termsOfService is a URL to the terms of service for this API.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	// +optional
	TermsOfService *string `json:"termsOfService,omitempty"`

	// contact is the contact information for this API. It may be an email
	// address, a URL, or a contact name. A value that parses as an email is
	// treated as an email address; a value that parses as a URL is treated as a
	// URL; any other value is treated as a contact name. Email takes precedence,
	// then URL. A URL must include a scheme (e.g. "https://") to be recognized
	// as one; a bare host like "www.example.com" is treated as a contact name.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +optional
	Contact *string `json:"contact,omitempty"`

	// license is the license name or identifier for this API (e.g., "MIT", "Apache-2.0").
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +optional
	License *string `json:"license,omitempty"`
}

const (
	// ApiProductConditionReady is the condition type for the top-level ApiProduct status.
	ApiProductConditionReady = "Ready"

	// Top-level ApiProduct reasons (used with ApiProductConditionReady).

	// ApiProductReasonAccepted indicates all versions were successfully processed.
	ApiProductReasonAccepted = "Accepted"
	// ApiProductReasonPartiallyAccepted indicates some versions are serving content while
	// others have errors or warnings.
	ApiProductReasonPartiallyAccepted = "PartiallyAccepted"
	// ApiProductReasonError indicates no versions were successfully processed.
	ApiProductReasonError = "Error"

	// ApiProductVersionConditionReady is the condition type for per-version status.
	ApiProductVersionConditionReady = "Ready"

	// Per-version reasons (used with ApiProductVersionConditionReady).

	// ApiProductVersionReasonAccepted indicates the version was fully and successfully stitched.
	ApiProductVersionReasonAccepted = "Accepted"
	// ApiProductVersionReasonWarning indicates the version produced a stitched doc but with
	// degraded output — one or more ApiDocs were excluded due to schema conflicts. The stitched
	// doc is still served; details are in the condition message.
	ApiProductVersionReasonWarning = "Warning"
	// ApiProductVersionReasonError indicates the version failed to produce a stitched doc.
	ApiProductVersionReasonError = "Error"
)

// ApiProductStatus defines the observed state of ApiProduct.
type ApiProductStatus struct {
	// conditions describe the current conditions of the ApiProduct.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// versions contains the status of each version defined in spec.versions.
	// +listType=map
	// +listMapKey=name
	// +optional
	Versions []ApiProductVersionStatus `json:"versions,omitempty"`
}

// ApiProductVersionStatus defines the observed state of a single ApiProduct version.
type ApiProductVersionStatus struct {
	// conditions describe the current conditions of this version.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +patchStrategy=merge
	// +patchMergeKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// name is the version name, matching spec.versions[].name.
	// +required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// generatedApiDocRef references the ApiDoc generated for this version.
	// This ApiDoc is managed by the controller and contains the stitched/projected
	// OpenAPI schema for this version. Nil if the version failed to process.
	// +optional
	GeneratedApiDocRef *GeneratedApiDocRef `json:"generatedApiDocRef,omitempty"`
}

// GeneratedApiDocRef is a reference to a generated ApiDoc.
type GeneratedApiDocRef struct {
	// name is the name of the generated ApiDoc.
	// +required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// namespace is the namespace of the generated ApiDoc.
	// +required
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`
}

// ApiProductList contains a list of ApiProduct resources
// +kubebuilder:object:root=true
type ApiProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiProduct `json:"items"`
}
