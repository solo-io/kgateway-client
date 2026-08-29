package enterprisekgateway

import (
	upstream "github.com/kgateway-dev/kgateway/v2/api/v1alpha1/kgateway"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// EntHeaderSanitization removes HTTP headers that are not permitted by the
// configured rules.
//
// Sanitization currently covers response headers only. Request headers are not
// yet supported, and trailers are not sanitized in either direction.
//
// Removals are observable two ways. The counters
// `solo.io.http.header_sanitization.removed` and
// `solo.io.http.header_sanitization.responses_sanitized` (responses, counted
// once each even when several stages strip from the same response) are always
// emitted. `removed` counts distinct header names rather than header lines: a
// header sent several times counts once, even though all of its values are
// dropped, because removal is keyed by name and one call drops every value under
// it. For auditing, the names of removed headers are appended to the
// dynamic metadata list
// `io.solo.http.header_sanitization:removed_response_headers`, which an access
// log can render with
// `%DYNAMIC_METADATA(io.solo.http.header_sanitization:removed_response_headers)%`.
// Only names are recorded, never values.
// +kubebuilder:validation:AtLeastOneOf=response
type EntHeaderSanitization struct {
	// Response sanitizes headers on the upstream response before they continue
	// toward the client. Response trailers are not affected.
	// +optional
	Response *HeaderSanitizationRules `json:"response,omitempty"`

	// FilterStage controls where in the HTTP filter chain sanitization runs.
	// Defaults to `{stage: Fault, predicate: Before}`.
	//
	// Envoy runs HTTP filters in order on the request path and in reverse order
	// on the response path, so the stage determines what else observes the
	// removed headers:
	//
	//   - `{stage: Route, predicate: After}` places the filter closest to the
	//     upstream, so it sanitizes as soon as the response arrives and no later
	//     filter in the HTTP connection manager chain sees the removed headers.
	//   - `{stage: Fault, predicate: Before}` (the default) places the filter
	//     closest to the client, so every other filter in that chain still sees
	//     the full header set but the client does not.
	//
	// `Before` and `After` are absolute within their stage: `Before` places this
	// filter ahead of every other filter at that stage and `After` places it
	// behind every one of them, rather than merely adjacent to the stage. So
	// `{stage: Route, predicate: After}` also runs ahead of, say,
	// `entTransformation`'s postRouting stage on the response path, even though
	// both resolve to the same stage.
	//
	// Both placements are in the HTTP connection manager chain. Filters in the
	// router's upstream HTTP filter chain run closer to the upstream than the
	// router itself, so on the response path they run before either placement
	// and always observe the full, unsanitized header set. An ext_proc whose own
	// filter stage is after route runs there, for example. Use this policy to
	// control what reaches the client and the rest of the HTTP connection
	// manager chain, not to hide headers from every other filter.
	//
	// Policies that resolve to different stages each get their own filter
	// instance, so several may be active on one route or gateway at once.
	// +optional
	FilterStage *upstream.FilterStageSpec `json:"filterStage,omitempty"`
}

// HeaderSanitizationRules selects which headers survive sanitization.
// +kubebuilder:validation:ExactlyOneOf=allowlist
type HeaderSanitizationRules struct {
	// Allowlist is the set of header names permitted through. Every header that
	// is not listed here, and not one of the built-in critical headers, is
	// removed. Matching is case-insensitive.
	//
	// A built-in set of critical standard headers is always permitted in
	// addition to this list and cannot be disabled. It covers message framing
	// (content-type, content-length, content-encoding, transfer-encoding, ...),
	// connection management (connection, upgrade, ...), caching (cache-control,
	// etag, expires, ...) and protocol correctness (location, www-authenticate,
	// retry-after, ...), plus all HTTP/2 pseudo-headers. Removing any of those
	// would produce a malformed rather than a merely reduced response.
	//
	// Note that `set-cookie` is deliberately NOT built in; list it explicitly if
	// the upstream sets cookies you need to keep.
	//
	// `server` is not built in either, but listing it achieves nothing: Envoy
	// sets that header after the filter chain has run, overwriting whatever is
	// there by default, so it is out of reach of this policy. Control it with the
	// listener's server header transformation instead. The same applies to
	// `date`, `via`, `proxy-status` and `x-request-id`, which Envoy may also add
	// after sanitization.
	//
	// An empty list is valid and removes every non-critical header.
	//
	// When two policies that target the same route and the same filter stage are
	// deep merged, their allowlists are unioned. Because a union is widening, a
	// lower-precedence policy can permit a header a higher-precedence policy
	// omitted; use a shallow merge strategy if you need strict override
	// semantics.
	// +optional
	// +kubebuilder:validation:MaxItems=256
	// +kubebuilder:listType=set
	Allowlist []gwv1.HTTPHeaderName `json:"allowlist,omitempty"`
}
