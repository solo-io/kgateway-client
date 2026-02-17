package shared

import (
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// RateLimitConfigRef selects the RateLimitConfig resource with the rate limit policy that you want to use. For more details, see the [RateLimitConfig reference in the Gloo Edge docs](https://docs.solo.io/gloo-edge/main/reference/api/github.com/solo-io/solo-apis/api/rate-limiter/v1alpha1/ratelimit.proto.sk/).
type RateLimitConfigRef struct {
	// Name is the name of the RateLimitConfig resource.
	// +required
	Name gwv1.ObjectName `json:"name"`

	// Namespace is the namespace of the RateLimitConfig resource.
	// If not set, defaults to the namespace of the policy from which the RateLimitConfig is referenced.
	// +optional
	Namespace *gwv1.Namespace `json:"namespace,omitempty"`
}

// AuthConfigRef selects the AuthConfig resource with the external auth policy that you want to use. For more details, see the [AuthConfig reference in the Gloo Edge docs](https://docs.solo.io/gloo-edge/main/reference/api/github.com/solo-io/gloo/projects/gloo/api/v1/enterprise/options/extauth/v1/extauth.proto.sk/#authconfig).
type AuthConfigRef struct {
	// Name is the name of the AuthConfig resource.
	// +required
	Name gwv1.ObjectName `json:"name"`

	// Namespace is the namespace of the AuthConfig resource.
	// If not set, defaults to the namespace of the policy from which the AuthConfig is referenced.
	// +optional
	Namespace *gwv1.Namespace `json:"namespace,omitempty"`
}

// WAFPolicyRef selects the WAFPolicy resource with the configuration that you want to use.
type WAFPolicyRef struct {
	// Name is the name of the WAFPolicy resource.
	// +required
	Name gwv1.ObjectName `json:"name"`

	// Namespace is the namespace of the WAFPolicy resource.
	// If not set, defaults to the namespace of the policy from which the WAFPolicy is referenced.
	// +optional
	Namespace *gwv1.Namespace `json:"namespace,omitempty"`
}
