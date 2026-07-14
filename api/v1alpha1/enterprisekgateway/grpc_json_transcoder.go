package enterprisekgateway

import (
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
)

// EntGrpcJsonTranscoder configures the gateway so that is transparently transcodes
// incoming HTTP/JSON requests into gRPC calls for the routes this policy is attached to
// and transcodes the gRPC responses back into JSON. The filter is only activated for routes/virtual
// hosts targeted by this policy via a per-route `typed_per_filter_config` override, so other routes
// on the same gateway are unaffected.
//
// Because Envoy applies the per-route transcoder config after route matching, routes using
// this policy must match against the original incoming HTTP path (this is equivalent to
// Envoy's `match_incoming_request_route=true` behavior).
//
// Note: The referenced gRPC backend must be configured for HTTP/2. Configure a `BackendConfigPolicy` with
// `http2ProtocolOptions` on the backend Service or use the `kubernetes.io/h2c` appProtocol on the Service port;
// otherwise transcoding will fail at runtime.
//
// +kubebuilder:validation:ExactlyOneOf=protoDescriptorBin;protoDescriptorConfigMap;disable
// +kubebuilder:validation:XValidation:rule="has(self.disable) || (has(self.services) && size(self.services) > 0)",message="services is required unless disable is set"
type EntGrpcJsonTranscoder struct {
	// ProtoDescriptorBin is the proto descriptor set (a compiled `.pb` `FileDescriptorSet`, e.g.
	// produced by `protoc --descriptor_set_out --include_imports`), supplied as base64-encoded
	// bytes. Use this for small or stable proto APIs where a self-contained policy is preferred.
	// Exactly one of `protoDescriptorBin`, `protoDescriptorConfigMap`, or `disable` must be set.
	//
	// The descriptor is limited to ~1 MiB. Because this is a byte field, the CRD MaxLength applies
	// to the base64-encoded form, so the limit is set to ceil(1 MiB * 4/3) = 1398104 encoded
	// characters (i.e. up to 1 MiB of raw descriptor bytes). For larger descriptors use
	// `protoDescriptorConfigMap`.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1398104
	ProtoDescriptorBin []byte `json:"protoDescriptorBin,omitempty"`

	// ProtoDescriptorConfigMap references a Kubernetes ConfigMap key that holds the proto
	// descriptor set. Prefer this when the descriptor is large, changes independently of the
	// policy, or is shared across policies. The control plane watches the ConfigMap and
	// rebuilds config on valid changes.
	// Exactly one of `protoDescriptorBin`, `protoDescriptorConfigMap`, or `disable` must be set.
	// +optional
	ProtoDescriptorConfigMap *ProtoDescriptorConfigMapRef `json:"protoDescriptorConfigMap,omitempty"`

	// Disable turns off gRPC-JSON transcoding for the targeted routes. Use this to opt a specific
	// route out of a transcoder policy inherited from a higher level in the config hierarchy (for
	// example a Gateway-wide policy). Exactly one of `protoDescriptorBin`,
	// `protoDescriptorConfigMap`, or `disable` must be set.
	// +optional
	Disable *upstreamshared.PolicyDisable `json:"disable,omitempty"`

	// Services is the list of fully-qualified gRPC service names (i.e. `package_name.service_name`,
	// for example `main.Bookstore`) that the transcoder should translate. Every name must exist
	// in the supplied proto descriptor. The descriptor may contain additional services that are
	// left untouched. Required unless `disable` is set.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=512
	// +kubebuilder:validation:items:Pattern=`^[A-Za-z_][A-Za-z0-9_.]*$`
	// +listType=set
	Services []string `json:"services,omitempty"`

	// PrintOptions controls how transcoded JSON responses are formatted.
	// +optional
	PrintOptions *GrpcJsonPrintOptions `json:"printOptions,omitempty"`

	// MatchIncomingRequestRoute keeps the incoming request route after outgoing headers have been
	// transformed to match the upstream gRPC service. Defaults to true.
	//
	// This policy activates the transcoder per-route (via a route-level filter config override),
	// so route matching runs against the original incoming HTTP path. Leaving this true is
	// required for the common route-attached case: with false, Envoy rewrites the path to the
	// gRPC `/package.Service/Method` form and re-routes, which no longer matches the HTTPRoute
	// this policy is attached to. Set false only for gateway/listener-wide attachments where you
	// have defined separate routes matching the transcoded gRPC paths.
	// +optional
	// +kubebuilder:default=true
	MatchIncomingRequestRoute *bool `json:"matchIncomingRequestRoute,omitempty"`

	// IgnoredQueryParameters is a list of query parameters to be ignored for transcoding method
	// mapping. By default the transcoder rejects a request if it contains unknown query parameters.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MinLength=1
	// +kubebuilder:validation:items:MaxLength=256
	// +listType=set
	IgnoredQueryParameters []string `json:"ignoredQueryParameters,omitempty"`

	// IgnoreUnknownQueryParameters ignores any query parameters that cannot be mapped to a
	// protobuf field, rather than rejecting the request. Use this when clients may send query
	// parameters you do not control. Defaults to false.
	// +optional
	IgnoreUnknownQueryParameters *bool `json:"ignoreUnknownQueryParameters,omitempty"`

	// AutoMapping routes methods that lack the `google.api.http` annotation using the default
	// `/package.Service/Method` path convention. Useful for third-party protos you do not control.
	// Defaults to false.
	// +optional
	AutoMapping *bool `json:"autoMapping,omitempty"`

	// ConvertGrpcStatus converts gRPC status trailers (and `google.rpc.Status` error details)
	// into an equivalent HTTP status code and JSON error body, so REST clients receive standard
	// HTTP error semantics. Defaults to false.
	// +optional
	ConvertGrpcStatus *bool `json:"convertGrpcStatus,omitempty"`
}

// ProtoDescriptorConfigMapRef references a proto descriptor set stored in a ConfigMap.
type ProtoDescriptorConfigMapRef struct {
	// Name of the ConfigMap.
	// +kubebuilder:validation:MinLength=1
	// +required
	Name string `json:"name"`

	// Namespace of the ConfigMap. Defaults to the namespace of this policy. Cross-namespace
	// references require a ReferenceGrant.
	// +optional
	Namespace *string `json:"namespace,omitempty"`

	// Key within the ConfigMap that holds the descriptor. Descriptor bytes stored under
	// `binaryData` are used verbatim; values stored under `data` are treated as base64-encoded.
	// If unset, the ConfigMap must contain exactly one entry, which is used.
	// +optional
	Key *string `json:"key,omitempty"`
}

// GrpcJsonPrintOptions controls JSON response formatting for transcoded responses.
// These map directly to Envoy's `GrpcJsonTranscoder.PrintOptions`.
type GrpcJsonPrintOptions struct {
	// AddWhitespace adds newlines, indentation and spaces to make the JSON easier to read.
	// Defaults to false.
	// +optional
	AddWhitespace *bool `json:"addWhitespace,omitempty"`

	// AlwaysPrintPrimitiveFields includes primitive fields set to their default values in the
	// response (they are omitted by default). Defaults to false.
	// +optional
	AlwaysPrintPrimitiveFields *bool `json:"alwaysPrintPrimitiveFields,omitempty"`

	// AlwaysPrintEnumsAsInts prints enum values as integers instead of their string names.
	// Defaults to false.
	// +optional
	AlwaysPrintEnumsAsInts *bool `json:"alwaysPrintEnumsAsInts,omitempty"`

	// PreserveProtoFieldNames uses the proto field name as-is instead of its lowerCamelCase JSON
	// name. Defaults to false.
	// +optional
	PreserveProtoFieldNames *bool `json:"preserveProtoFieldNames,omitempty"`
}
