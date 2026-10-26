package enterprisekgateway

// TransformationRequestMatcher configures the matcher to match against the request.
// +kubebuilder:validation:AtMostOneOf=prefix;path;regex;connect
type TransformationRequestMatcher struct {
	// Prefix configures the prefix rule meaning that the prefix must
	// match the beginning of the *:path* header.
	// Max length is following https://gateway-api.sigs.k8s.io/reference/spec/#httppathmatch
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +optional
	Prefix *string `json:"prefix,omitempty"`

	// Path configures the exact path rule meaning that the path must
	// exactly match the *:path* header once the query string is removed.
	// Max length is following https://gateway-api.sigs.k8s.io/reference/spec/#httppathmatch
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +optional
	Path *string `json:"path,omitempty"`

	// Regex configures the route regular expression rule meaning that the
	// regex must match the *:path* header once the query string is removed. The entire path
	// (without the query string) must match the regex. The rule will not match if only a
	// subsequence of the *:path* header matches the regex.
	// +optional
	Regex *RegexMatcher `json:"regex,omitempty"`

	// Connect configures the matcher to only match CONNECT requests.
	// Note that this will not match HTTP/2 upgrade-style CONNECT requests
	// (WebSocket and the like) as they are normalized in Envoy as HTTP/1.1 style
	// upgrades.
	// This is the only way to match CONNECT requests for HTTP/1.1. For HTTP/2,
	// where CONNECT requests may have a path, the path matchers will work if
	// there is a path present.
	// +optional
	Connect *bool `json:"connect,omitempty"`

	// CaseSensitive indicates that prefix/path matching should be case-insensitive. The default
	// is true.
	// +kubebuilder:default=true
	// +optional
	CaseSensitive *bool `json:"caseSensitive,omitempty"`

	// Specifies a set of headers that the route should match on. The router will
	// check the request's headers against all the specified headers in the route
	// config. A match will happen if all the headers in the route are present in
	// the request with the same values (or based on presence if the value field
	// is not in the config).
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Headers []TransformationHeaderMatcher `json:"headers,omitempty"`

	// Specifies a set of URL query parameters on which the route should
	// match. The router will check the query string from the *path* header
	// against all the specified query parameters. If the number of specified
	// query parameters is nonzero, they all must match the *path* header's
	// query string for a match to occur.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	QueryParameters []QueryParameterMatcher `json:"queryParameters,omitempty"`

	// If specified, only gRPC requests will be matched. The router will check
	// that the content-type header has a application/grpc or one of the various
	// application/grpc+ values.
	// +optional
	Grpc *bool `json:"grpc,omitempty"`

	// If specified, the client tls context will be matched against the defined
	// match options.
	// +optional
	TlsContext *TlsContextMatchOptions `json:"tlsContext,omitempty"`

	// HTTP Method/Verb(s) to match on. If none specified, the matcher will ignore the HTTP Method
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Methods []string `json:"methods,omitempty"`
}

// TransformationHeaderMatcher configures the header matching to apply.
type TransformationHeaderMatcher struct {
	// Specifies the name of the header in the request.
	// +required
	Name string `json:"name"`
	// Specifies the value of the header. If the value is absent a request that
	// has the name header will match, regardless of the header's value.
	// +optional
	Value *string `json:"value,omitempty"`
	// Specifies whether the header value should be treated as regex or not.
	// +optional
	Regex *bool `json:"regex,omitempty"`
	// If set to true, the result of the match will be inverted. Defaults to false.
	//
	// Examples:
	// * name=foo, invertMatch=true: matches if no header named `foo` is present
	// * name=foo, value=bar, invertMatch=true: matches if no header named `foo` with value `bar` is present
	// * name=foo, value=`\d{3}`, regex=true, invertMatch=true: matches if no header named `foo` with a value consisting of three integers is present
	// +optional
	InvertMatch *bool `json:"invertMatch,omitempty"`
}

// RegexMatcher based on https://github.com/envoyproxy/envoy/blob/4453ce1f809ec502fb2cbe0363cf5c6a971f3836/api/envoy/type/matcher/regex.proto#L19
type RegexMatcher struct {
	// The regex match string. The string must be supported by the configured engine.
	// +required
	Regex string `json:"regex"`
}

// RangeMatch configures the header match to be performed based on range.
type RangeMatch struct {
	// RangeMatch will configure the header match to be performed based on range.
	// The rule will match if the request header value is within this range.
	// The entire request header value must represent an integer in base 10 notation: consisting of
	// an optional plus or minus sign followed by a sequence of digits. The rule will not match if
	// the header value does not represent an integer. Match will fail for empty values, floating
	// point numbers or if only a subsequence of the header value is an integer.
	//
	// Examples:
	//
	//   - For range [-10,0), route will match for header value -1, but not for 0, "somestring", 10.9,
	//     "-1somestring"
	// +required
	RangeMatch int64 `json:"rangeMatch"`
}

// StringMatch configures the string matching to apply.
// +kubebuilder:validation:ExactlyOneOf=exact;prefix;suffix;regex
type StringMatch struct {
	// The input string must match exactly the string specified here.
	//
	// Examples:
	//
	// * *abc* only matches the value *abc*.
	// +optional
	Exact *string `json:"exact,omitempty"`

	// The input string must have the prefix specified here.
	// Note: empty prefix is not allowed, please use regex instead.
	//
	// Examples:
	//
	// * *abc* matches the value *abc.xyz*
	// +optional
	Prefix *string `json:"prefix,omitempty"`

	// The input string must have the suffix specified here.
	// Note: empty prefix is not allowed, please use regex instead.
	//
	// Examples:
	//
	// * *abc* matches the value *xyz.abc*
	// +optional
	Suffix *string `json:"suffix,omitempty"`

	// The input string must match the regular expression specified here.
	// +optional
	Regex *RegexMatcher `json:"regex,omitempty"`

	// If true, indicates the exact/prefix/suffix matching should be case-insensitive. This has no
	// effect for the regex match.
	// For example, the matcher *data* will match both input string *Data* and *data* if set to true.
	// +optional
	IgnoreCase *bool `json:"ignoreCase,omitempty"`
}

// QueryParameterMatcher configures the query parameter matching to apply.
type QueryParameterMatcher struct {
	// Specifies the name of a key that must be present in the requested
	// *path*'s query string.
	// +required
	Name string `json:"name"`
	// Specifies the value of the key. If the value is absent, a request
	// that contains the key in its query string will match, whether the
	// key appears with a value (e.g., "?debug=true") or not (e.g., "?debug")
	// +optional
	Value *string `json:"value,omitempty"`
	// Specifies whether the query parameter value is a regular expression.
	// Defaults to false. The entire query parameter value (i.e., the part to
	// the right of the equals sign in "key=value") must match the regex.
	// E.g., the regex "\d+$" will match "123" but not "a123" or "123a".
	// +optional
	Regex *bool `json:"regex,omitempty"`
}

// TlsContextMatchOptions configures the TLS context match options.
type TlsContextMatchOptions struct {
	// If specified, the route will match against whether a certificate is presented.
	// If not specified, certificate presentation status (true or false) will not be considered when route matching.
	// +optional
	Presented *bool `json:"presented,omitempty"`
	// If specified, the route will match against whether a certificate is validated.
	// If not specified, certificate validation status (true or false) will not be considered when route matching.
	// +optional
	Validated *bool `json:"validated,omitempty"`
}
