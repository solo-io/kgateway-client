package enterprisekgateway

import (
	upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// StagedJWT allows for configuring JWT authentication at various stages of request processing
// +kubebuilder:validation:XValidation:rule="has(self.afterExtAuth) || has(self.beforeExtAuth)",message="for staged JWT usage, at least one stage must be set"
type StagedJWT struct {
	// JWT configuration to be enforced after external auth has been processed (if it is present).
	// Note: this is not currently supported for agentgateway.
	// +optional
	AfterExtAuth *EntJWT `json:"afterExtAuth,omitempty"`

	// JWT configuration to be enforced before external auth has been processed.
	// +optional
	BeforeExtAuth *EntJWT `json:"beforeExtAuth,omitempty"`
}

// EntJWT defines a set of providers used for JWT authentication (and an optional validation policy for these providers) or
// the ability to disable JWT authentication and verification.
// +kubebuilder:validation:XValidation:rule="!has(self.providers) || !has(self.disable)",message="providers can not be set if disable is set"
// +kubebuilder:validation:XValidation:rule="!has(self.validationPolicy) || !has(self.disable)",message="validationPolicy can not be set if disable is set"
type EntJWT struct {
	// Providers maps a provider name to a JWT provider, configuring a way to authenticate JWTs.
	// If specified, multiple providers will be `OR`-ed together and will allow validation to any of the providers.
	// Note: agentgateway only supports a single provider. If more than one provider is specified,
	// the first provider found with a local JWKS will be used,
	// but order is not guaranteed to be respected due to the map type.
	// +optional
	// +kubebuilder:validation:MaxProperties=32
	Providers map[string]JWTProvider `json:"providers,omitempty"`

	// Configure how JWT validation works, with the flexibility to handle requests with missing or invalid JWTs.
	// By default, after applying a JWT policy, only requests that have been authenticated with a valid JWT are allowed.
	// +optional
	ValidationPolicy *JwtValidationPolicy `json:"validationPolicy,omitempty"`

	// Disable JWT authentication for this policy scope.
	// Note: this is not currently supported for agentgateway.
	// +optional
	Disable *upstreamshared.PolicyDisable `json:"disable,omitempty"`
}

// JWTProvider defines configuration for how a JWT should be authenticated and verified.
type JWTProvider struct {
	// The source for the keys to validate JWTs.
	// +required
	JWKS JWKS `json:"jwks"`

	// An incoming JWT must have an 'aud' claim and it must be in this list.
	// +optional
	Audiences []string `json:"audiences,omitempty"`

	// Issuer of the JWT. the 'iss' claim of the JWT must match this.
	// +optional
	Issuer *string `json:"issuer,omitempty"`

	// Where to find the JWT of the current provider.
	// Note: agentgateway does not support token source configuration.
	// +optional
	TokenSource *TokenSource `json:"tokenSource,omitempty"`

	// Should the token forwarded upstream. If false, the header containing the token will be removed.
	// If omitted, the default behavior is to remove the token and not forward
	// +optional
	KeepToken *bool `json:"keepToken,omitempty"`

	// What claims should be copied to upstream headers.
	// Note: agentgateway does not support claimsToHeaders configuration.
	// +optional
	ClaimsToHeaders []ClaimToHeader `json:"claimsToHeaders,omitempty"`

	// Used to verify time constraints, such as `exp` and `npf`. If omitted, defaults to 60s
	// Note: agentgateway does not support clockSkewSeconds configuration.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ClockSkewSeconds *int32 `json:"clockSkewSeconds,omitempty"`

	// When this field is set, the specified value is used as the key in DynamicMetadata to store the JWT failure status code
	// and message under that key. This field is particularly useful when logging the failure status.
	// Note: agentgateway does not support attachFailedStatusToMetadata configuration.
	// For example, if the value of `attach_failed_status_to_metadata` is 'custom_auth_failure_status' then
	// the failure status can be accessed in the access log as '%DYNAMIC_METADATA(envoy.filters.http.jwt_authn:custom_auth_failure_status)'
	// Note: status code and message can be individually accessed as '%DYNAMIC_METADATA(envoy.filters.http.jwt_authn:custom_auth_failure_status.code)' and '%DYNAMIC_METADATA(envoy.filters.http.jwt_authn:custom_auth_failure_status.message)' respectively.
	// +optional
	AttachFailedStatusToMetadata *string `json:"attachFailedStatusToMetadata,omitempty"`
}

// Allows copying verified claims to headers sent upstream
type ClaimToHeader struct {
	// Claim name. for example, "sub"
	// +required
	Claim string `json:"claim"`

	// The header the claim will be copied to. for example, "x-sub".
	// +required
	Header string `json:"header"`

	// If the header exists, append to it (true), or overwrite it (false).
	// If omitted, will default to false.
	// +optional
	Append *bool `json:"append,omitempty"`
}

// JWKS (JSON Web Key Set) configures how to fetch the public key used for JWT verification.
// +kubebuilder:validation:ExactlyOneOf=local;remote
type JWKS struct {
	// Local is used when JWKS is local to the proxy, such as an inline string definition.
	// +optional
	Local *LocalJWKS `json:"local,omitempty"`

	// Remote is used when the JWKS should be fetched from a remote host
	// Note: agentgateway does not support remote JWKS configuration.
	// +optional
	Remote *RemoteJWKS `json:"remote,omitempty"`
}

// LocalJWKS contains configuration for JWKS that are locally available to the proxy
type LocalJWKS struct {
	// Inline key. this can be json web key, key-set or PEM format.
	// +required
	Key string `json:"key"`
}

type RemoteJWKS struct {
	// The url used when accessing the upstream for Json Web Key Set.
	// This is used to correctly set the host and path in the JWKS HTTP request.
	// E.g. https://example.com/.well-known/jwks.json
	// +kubebuilder:validation:Pattern=`^(http|https):\/\/[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*(:\d+)?\/.*$`
	// +required
	Url string `json:"url"`

	// The Backend representing the Json Web Key Set server
	// +required
	BackendRef gwv1.BackendRef `json:"backendRef"`

	// Duration after which the cached JWKS should be expired.
	// If not specified, default cache duration is 5 minutes.
	// +optional
	// +kubebuilder:validation:XValidation:rule="matches(self, '^([0-9]{1,5}(h|m|s|ms)){1,4}$')",message="invalid duration value"
	// +kubebuilder:validation:XValidation:rule="duration(self) >= duration('1ms')",message="cacheDuration must be at least 1ms."
	CacheDuration *metav1.Duration `json:"cacheDuration,omitempty"`

	// Fetch Jwks asynchronously in the main thread before the listener is activated.
	// Fetched Jwks can be used by all worker threads.
	//
	// If this feature is not enabled:
	//   - The Jwks is fetched on-demand when the requests come. During the fetching, first
	//     few requests are paused until the Jwks is fetched.
	//   - Each worker thread fetches its own Jwks since Jwks cache is per worker thread.
	//
	// If this feature is enabled:
	//   - Fetched Jwks is done in the main thread before the listener is activated. Its fetched
	//     Jwks can be used by all worker threads. Each worker thread doesn't need to fetch its own.
	//   - Jwks is ready when the requests come, not need to wait for the Jwks fetching.
	// +optional
	AsyncFetch *JwksAsyncFetch `json:"asyncFetch,omitempty"`
}

// Fetch Jwks asynchronously in the main thread when the filter config is parsed.
// The listener is activated only after the Jwks is fetched.
// When the Jwks is expired in the cache, it is fetched again in the main thread.
// The fetched Jwks from the main thread can be used by all worker threads.
type JwksAsyncFetch struct {
	// If false, the listener is activated after the initial fetch is completed.
	// The initial fetch result can be either successful or failed.
	// If true, it is activated without waiting for the initial fetch to complete.
	// Default is false.
	// +optional
	FastListener *bool `json:"fastListener,omitempty"`
}

// Describes the location of a JWT token
type TokenSource struct {
	// Try to retrieve token from these headers
	// +optional
	Headers []TokenSourceHeaderSource `json:"headers,omitempty"`

	// Try to retrieve token from these query params
	// +optional
	QueryParams []string `json:"queryParams,omitempty"`
}

// Describes how to retrieve a JWT from a header
type TokenSourceHeaderSource struct {
	// The name of the header. for example, "authorization"
	// +required
	Header string `json:"header"`

	// Prefix before the token. for example, "Bearer "
	// +optional
	Prefix *string `json:"prefix,omitempty"`
}

// +kubebuilder:validation:Enum=RequireValid;AllowMissing;AllowMissingOrFailed
type JwtValidationPolicy string

const (
	// Default value. Allow only requests that authenticate with a valid JWT to succeed.
	ValidationPolicyRequireValid JwtValidationPolicy = "RequireValid"

	// Allow requests to succeed even if JWT authentication is missing, but fail when an invalid JWT token is presented.
	// You might use this setting when later steps depend on input from the JWT.
	// For example, you might add claims from the JWT to request headers with the claimsToHeaders field.
	// As such, you may want to make sure that any provided JWT is valid. If not, the request fails,
	// which informs the requester that their JWT is not valid.
	// Requests without a JWT, however, still succeed and skip JWT validation.
	ValidationPolicyAllowMissing JwtValidationPolicy = "AllowMissing"

	// Allow requests to succeed even when a JWT is missing or JWT verification fails.
	// For example, you might apply multiple policies to your routes so that requests can authenticate with either a
	// JWT or another method such as external auth. Use this value
	// to allow a failed JWT auth request to pass through to the other authentication method.
	ValidationPolicyAllowMissingOrFailed JwtValidationPolicy = "AllowMissingOrFailed"
)
