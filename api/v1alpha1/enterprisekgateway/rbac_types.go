package enterprisekgateway

import upstreamshared "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/shared"

// EntRBAC defines RBAC configuration.
// +kubebuilder:validation:ExactlyOneOf=disable;policies
type EntRBAC struct {
	// Disable is used to explicitly disable RBAC checks for the scope of this policy.
	// This is useful to allow access to static resources/login page without RBAC checks.
	// +optional
	Disable *upstreamshared.PolicyDisable `json:"disable,omitempty"`

	// Policies maps a policy name to an RBAC policy to apply.
	// +optional
	Policies map[string]RBACPolicy `json:"policies,omitempty"`
}

type RBACPolicy struct {
	// Principals in this policy.
	// +kubebuilder:validation:MinItems=1
	// +required
	Principals []RBACPrincipal `json:"principals"`

	// Permissions granted to the principals.
	// +optional
	Permissions *RBACPermissions `json:"permissions,omitempty"`

	// The delimiter to use when specifying nested claim names within principals.
	// Default is an empty string, which disables nested claim functionality.
	// This is commonly set to `.`, allowing for nested claim names of the form
	// `parent.child.grandchild`
	// +optional
	NestedClaimDelimiter *string `json:"nestedClaimDelimiter,omitempty"`
}

// An RBAC principal - the identity entity (usually a user or a service account).
type RBACPrincipal struct {
	// JWTPrincipal references a principal from JWT authentication.
	// +required
	JWTPrincipal RBACJWTPrincipal `json:"jwtPrincipal"`
}

// A JWT principal. To use this, JWT authentication MUST be configured as well.
type RBACJWTPrincipal struct {
	// Set of claims that make up this principal. Commonly, the 'iss' and 'sub' or 'email' claims are used.
	// If you specify the path for a nested claim, such as 'parent.child.foo', you must also specify
	// a non-empty string value for the `nested_claim_delimiter` field in the Policy.
	// +required
	Claims map[string]string `json:"claims"`

	// Verify that the JWT came from a specific provider. This usually can be left empty
	// and a provider will be chosen automatically.
	// +optional
	// +kubebuilder:validation:MinLength=1
	Provider *string `json:"provider,omitempty"`

	// The matcher to use when evaluating this principal. If omitted, exact string comparison (ExactString) is used.
	// +optional
	// +kubebuilder:validation:Enum=ExactString;Boolean;ListContains
	Matcher *RBACJWTPrincipalClaimMatcher `json:"matcher,omitempty"`
}

type RBACJWTPrincipalClaimMatcher string

const (
	// The JWT claim value is a string that exactly matches the value.
	JwtPrincipalClaimMatcherExactString RBACJWTPrincipalClaimMatcher = "ExactString"
	// The JWT claim value is a boolean that matches the value.
	JwtPrincipalClaimMatcherBoolean RBACJWTPrincipalClaimMatcher = "Boolean"
	// The JWT claim value is a list that contains a string that exactly matches the value.
	JwtPrincipalClaimMatcherListContains RBACJWTPrincipalClaimMatcher = "ListContains"
)

// What permissions should be granted. An empty field means allow-all.
// If more than one field is added, all of them need to match.
type RBACPermissions struct {
	// Paths that have this prefix will be allowed.
	// +optional
	// +kubebuilder:validation:MinLength=1
	PathPrefix *string `json:"pathPrefix,omitempty"`

	// What http methods (GET, POST, ...) are allowed.
	// +optional
	Methods []string `json:"methods,omitempty"`
}
