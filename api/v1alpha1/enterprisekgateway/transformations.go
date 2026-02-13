package enterprisekgateway

// TransformationExtractMode represents the mode of operation for the extraction, which configures how the tranformation
// will extract the content of a specified capturing group.
// +kubebuilder:validation:Enum=Extract;SingleReplace;ReplaceAll
type TransformationExtractMode string

const (
	// ModeExtract configures the transformation to extract the content of a specified capturing group. In this mode,
	// `subgroup` selects the n-th capturing group, which represents the value that
	// you want to extract.
	ModeExtract TransformationExtractMode = "Extract"

	// ModeSingleReplace configures the transformation to replace the content of a specified capturing group. In this mode, `subgroup` selects the
	// n-th capturing group, which represents the value that you want to replace with
	// the string provided in `replacementText`.
	// Note: `replacementText` must be set for this mode.
	ModeSingleReplace TransformationExtractMode = "SingleReplace"

	// ModeReplaceAll configures the transformation to replace all regex matches with the value provided in `replacementText`.
	// Note: `replacementText` must be set for this mode.
	// Note: The configuration fails if `subgroup` is set to a non-zero value.
	// Note: restrictions on the regex are different for this mode. See the regex field for more details.
	ModeReplaceAll TransformationExtractMode = "ReplaceAll"
)

// RequestBodyParse determines how the body will be parsed.
type RequestBodyParse string

const (
	// ParseAsJson configures the transformation to attempt to parse the request/response body as JSON
	ParseAsJson RequestBodyParse = "ParseAsJson"
	// DontParse configures the transformation request/response body will be treated as plain text
	DontParse RequestBodyParse = "DontParse"
)

// EntTransformation defines the Enterprise transformation configuration.
type EntTransformation struct {
	// Stages defines the transformations run at different stages of the filter chain.
	// +optional
	Stages *StagedTransformations `json:"stages,omitempty"`

	// AWSLambda defines the AWS Lambda transformation configuration.
	// +optional
	AWSLambda *AWSLambdaTransformation `json:"awsLambda,omitempty"`
}

// TODO(npolshak): Add support for XSLT as part of https://github.com/solo-io/gloo-gateway/issues/106
// TODO(npolshak): Add support for WrapAsAPIGateway as part of https://github.com/solo-io/gloo-gateway/issues/78
// TODO(npolshak): Add support for DLP as part of https://github.com/solo-io/gloo-gateway/issues/33

// Transformation defines a transformation that can be applied to requests or responses.
// +kubebuilder:validation:ExactlyOneOf=template;headerBody
type Transformation struct {
	// Template specifies a template-based transformation.
	// +optional
	Template *TransformationTemplate `json:"template,omitempty"`

	// HeaderBody specifies a header and body transformation.
	// +optional
	HeaderBody *HeaderBodyTransform `json:"headerBody,omitempty"`
}

// EscapeCharactersBehavior defines how to handle characters that need to be escaped in JSON.
// +kubebuilder:validation:Enum=Escape;DontEscape
type EscapeCharactersBehavior string

const (
	// EscapeCharactersEscape always escapes characters that need to be escaped in JSON
	EscapeCharactersEscape EscapeCharactersBehavior = "Escape"
	// EscapeCharactersDontEscape never escapes characters
	EscapeCharactersDontEscape EscapeCharactersBehavior = "DontEscape"
)

// TransformationTemplate defines a transformation template.
type TransformationTemplate struct {
	// AdvancedTemplates determines whether to use JSON pointer notation instead of dot notation.
	// If set to true, use JSON pointer notation (e.g. "time/start") instead of
	// dot notation (e.g. "time.start") to access JSON elements. Defaults to
	// false.
	//
	// Please note that, if set to 'true', you will need to use the `extraction`
	// function to access extractors in the template (e.g. "{{ extraction("my_extractor") }}").
	// If the default value of 'false' is used, extractors will simply be available by their name (e.g. "{{ my_extractor }}").
	// +optional
	AdvancedTemplates *bool `json:"advancedTemplates,omitempty"`

	// Extractors use this attribute to extract information from the request. It consists of
	// a map of strings to extractors. The extractor will define which
	// information will be extracted, while the string key will provide the
	// extractor with a name. You can reference extractors by their name in
	// templates, e.g. "{{ my-extractor }}" will render to the value of the
	// "my-extractor" extractor.
	// +optional
	// +kubebuilder:validation:MaxProperties=32
	Extractors map[string]Extraction `json:"extractors,omitempty"`

	// Headers configures the transform request/response headers. It consists of a
	// map of strings to templates. The string key determines the name of the
	// resulting header, the rendered template will determine the value. Any existing
	// headers with the same header name will be replaced by the transformed header.
	// If a header name is included in `headers` and `headersToAppend`, it will first
	// be replaced the template in `headers`, then additional header values will be appended
	// by the templates defined in `headersToAppend`.
	// For example, the following header transformation configuration:
	//
	// ```yaml
	//
	//	headers:
	//	  x-header-one: {"text": "first {{inja}} template"}
	//	  x-header-one: {"text": "second {{inja}} template"}
	//	headersToAppend:
	//	  - key: x-header-one
	//	    value: {"text": "first appended {{inja}} template"}
	//	  - key: x-header-one
	//	    value: {"text": "second appended {{inja}} template"}
	//
	// ```
	// will result in the following headers on the HTTP message:
	//
	// ```
	// x-header-one: first inja template
	// x-header-one: first appended inja template
	// x-header-one: second appended inja template
	// ```
	// +optional
	// +kubebuilder:validation:MaxProperties=32
	Headers map[string]InjaTemplate `json:"headers,omitempty"`

	// HeadersToAppend configures the transform request/response headers. It consists of
	// an array of string/template objects. Use this attribute to define multiple
	// templates for a single header. Header template(s) defined here will be appended to any
	// existing headers with the same header name, not replace existing ones.
	// See `headers` documentation to see an example of usage.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	HeadersToAppend []HeaderToAppend `json:"headersToAppend,omitempty"`

	// HeadersToRemove is configured to remove headers from requests. If a header is present multiple
	// times, all instances of the header will be removed.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	HeadersToRemove []string `json:"headersToRemove,omitempty"`

	// BodyTransformation specifies how to transform the body.
	// +optional
	BodyTransformation *BodyTransformation `json:"bodyTransformation,omitempty"`

	// ParseBodyBehavior determines how the body will be parsed. Defaults to ParseAsJson.
	// +kubebuilder:validation:Enum=ParseAsJson;DontParse
	// +kubebuilder:default=ParseAsJson
	// +optional
	ParseBodyBehavior *RequestBodyParse `json:"parseBodyBehavior,omitempty"`

	// IgnoreErrorOnParse determines whether Envoy should throw an exception if body parsing fails.
	// +optional
	IgnoreErrorOnParse *bool `json:"ignoreErrorOnParse,omitempty"`

	// DynamicMetadataValues defines Envoy Dynamic Metadata entries.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	DynamicMetadataValues []DynamicMetadataValue `json:"dynamicMetadataValues,omitempty"`

	// EscapeCharacters configures the Inja behavior when rendering strings which contain
	// characters that would need to be escaped to be valid JSON. Note that this
	// sets the behavior for the entire transformation. Use raw_strings function
	// for fine-grained control within a template.
	// +optional
	EscapeCharacters *EscapeCharactersBehavior `json:"escapeCharacters,omitempty"`

	// SpanTransformer defines a span transformer for modifying trace spans.
	// +optional
	SpanTransformer *SpanTransformer `json:"spanTransformer,omitempty"`
}

type OverridableTemplate struct {
	// Template to render
	// +required
	Tmpl InjaTemplate `json:"tmpl"`
	// If set to true, the template will be set even if the rendered value is empty.
	// +optional
	OverrideEmpty *bool `json:"overrideEmpty,omitempty"`
}

// Extraction is used to define extractions to extract information from the request/response.
// The extracted information can then be referenced in template fields.
// +kubebuilder:validation:AtMostOneOf=body;header
type Extraction struct {
	// ExtractionBody specifies extracting information from the request/response body.
	// +optional
	ExtractionBody *bool `json:"body,omitempty"`

	// ExtractionHeader specifies extracting information from headers.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	ExtractionHeader *string `json:"header,omitempty"`

	// Regex specifies the regular expression used for matching against the source content.
	//   - In Extract mode, the entire source must match the regex. `subgroup` selects the n-th capturing group,
	//     which determines the part of the match that you want to extract. If the regex does not match the source,
	//     the result of the extraction will be an empty value.
	//   - In SingleReplace mode, the regex also needs to match the entire source. `subgroup` selects the n-th capturing group
	//     that is replaced with the content of `replacementText`. If the regex does not match the source, the result
	//     of the replacement will be the source itself.
	//   - In ReplaceAll mode, the regex is applied repeatedly to find all occurrences within the source that match.
	//     Each matching occurrence is replaced with the value in `replacementText`. In this mode, the configuration is rejected
	//     if `subgroup` is set. If the regex does not match the source, the result of the replacement will be the source itself.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +required
	Regex string `json:"regex"`

	// Subgroup is used to determine the group that you want to select if your regex contains capturing groups. Defaults to 0.
	// If set in `Extract` and `SingleReplace` modes, the subgroup represents the capturing
	// group that you want to extract or replace in the source.
	// The configuration is rejected if you set subgroup to a non-zero value when using the `REPLACE_ALL` mode.
	// +optional
	// +kubebuilder:validation:Minimum=0
	Subgroup *int32 `json:"subgroup,omitempty"`

	// ReplacementText is used to format the substitution for matched sequences in
	// an input string. This value is only legal in `SingleReplace` and `REPLACE_ALL` modes.
	// - In `SingleReplace` mode, the `subgroup` selects the n-th capturing group, which represents
	// the value that you want to replace with the string provided in `replacementText`.
	// - In `REPLACE_ALL` mode, each sequence that matches the specified regex in the input is
	// replaced with the value in`replacementText`.
	//
	//	The `replacementText` can include special syntax, such as $1, $2, etc., to refer to
	//
	// capturing groups within the regular expression.
	//
	//	The value that is specified in `replacementText` is treated as a string, and is passed
	//
	// to `std::regex_replace` as the replacement string.
	//
	//	For more information, see https://en.cppreference.com/w/cpp/regex/regex_replace.
	// +optional
	ReplacementText *string `json:"replacementText,omitempty"`

	// Mode defines the mode of operation for the extraction.
	// Defaults to Extract.
	// +optional
	// +kubebuilder:default=Extract
	Mode *TransformationExtractMode `json:"mode,omitempty"`
}

// HeaderToAppend defines a header-template pair for appending headers.
type HeaderToAppend struct {
	// Key specifies the header name.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +required
	Key string `json:"key"`

	// Value specifies the template to apply to the header value.
	// +required
	Value InjaTemplate `json:"value"`
}

// InjaTemplate defines an [Inja template](https://github.com/pantor/inja) that will be
// rendered by Gloo. In addition to the core template functions, the Gloo
// transformation filter defines the following custom functions:
// - header(header_name): returns the value of the header with the given name.
// - extraction(extractor_name): returns the value of the extractor with the
// given name.
// - env(env_var_name): returns the value of the environment variable with the
// given name.
// - body(): returns the request/response body.
// - context(): returns the base JSON context (allowing for example to range on
// a JSON body that is an array).
// - request_header(header_name): returns the value of the request header with
// the given name. Use this option when you want to include request header values in response
// transformations.
// - base64_encode(string): encodes the input string to base64.
// - base64_decode(string): decodes the input string from base64.
// - substring(string, start_pos, substring_len): returns a substring of the
// input string, starting at `start_pos` and extending for `substring_len`
// characters. If no `substring_len` is provided or `substring_len` is <= 0, the
// substring extends to the end of the input string.
type InjaTemplate string

// DynamicMetadataValue defines an [Envoy Dynamic
// Metadata](https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata)
// entry.
type DynamicMetadataValue struct {
	// MetadataNamespace specifies the metadata namespace. Defaults to the filter namespace.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	MetadataNamespace *string `json:"metadataNamespace,omitempty"`

	// Key specifies the metadata key.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +required
	Key string `json:"key"`

	// Value specifies the template that determines the metadata value.
	// +required
	Value InjaTemplate `json:"value"`

	// JsonToProto determines whether to parse the rendered value as a proto Struct message.
	// +optional
	JsonToProto *bool `json:"jsonToProto,omitempty"`
}

// SpanTransformer defines a span transformer for modifying trace spans.
type SpanTransformer struct {
	// Name specifies a template that sets the span name.
	// +required
	Name InjaTemplate `json:"name"`
}

// HeaderBodyTransform defines a header and body transformation.
type HeaderBodyTransform struct {
	// AddRequestMetadata determines whether to add request metadata to the body.
	// When transforming a request, setting this to true will additionally add "queryString",
	// "queryStringParameters", "multiValueQueryStringParameters", "httpMethod", "path",
	// and "multiValueHeaders" to the body.
	// +optional
	AddRequestMetadata *bool `json:"addRequestMetadata,omitempty"`
}

// StagedTransformations configures transformations to apply for different stages of the filter chain.
type StagedTransformations struct {
	// Early transformations happen before most other options (Like Auth and Rate Limit).
	// +optional
	Early *RequestResponseTransformations `json:"early,omitempty"`

	// Regular transformations happen after Auth and Rate limit decisions have been made.
	// +optional
	Regular *RequestResponseTransformations `json:"regular,omitempty"`

	// PostRouting happen during the router filter chain. This is important for a number of reasons
	// 1. Retries re-trigger this filter, which might impact performance.
	// 2. It is the only point where endpoint metadata is available.
	// 3. `clearRouteCache` does NOT work in this stage as the routing decision is already made.
	// +optional
	PostRouting *RequestResponseTransformations `json:"postRouting,omitempty"`

	// TODO(npolshak): support transformation inheritance (https://github.com/solo-io/gloo-gateway/issues/153)

	// When enabled, log request/response body and headers before and after all transformations defined here are applied.\
	// This overrides the logRequestResponseInfo field in the Transformation message.
	// +optional
	LogRequestResponseInfo *bool `json:"logRequestResponseInfo,omitempty"`

	// EscapeCharacters configures the Inja behavior when rendering strings which contain
	// characters that would need to be escaped to be valid JSON. Note that this
	// sets the behavior for all staged transformations configured here. This setting
	// can be overridden per-transformation using the field `escapeCharacters` on
	// the TransformationTemplate.
	// +optional
	EscapeCharacters *EscapeCharactersBehavior `json:"escapeCharacters,omitempty"`
}

// RequestResponseTransformations configures transformations to apply on the request and response.
type RequestResponseTransformations struct {
	// Requests configures transformations to apply on the request. The first request that matches will apply.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Requests []RequestMatcher `json:"requests,omitempty"`

	// Responses configures transformations to apply on the response. The first response transformation that
	// matches will apply.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Responses []ResponseMatcher `json:"responses,omitempty"`
}

// RequestMatcher configures transformations to apply on the request.
type RequestMatcher struct {
	// Matcher defines the request matching parameter. Only when the match is satisfied, the "requires" field will
	// apply.
	//
	//
	// Matches define conditions used for matching the rule against incoming
	// HTTP requests. Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	//
	// For example, take the following matches configuration:
	//
	// ```
	// matches:
	// - path:
	//	   value: "/foo"
	//	 headers:
	//	 - name: "version"
	//	   value "v1"
	// - path:
	//	   value: "/v2/foo"
	// ```
	// For a request to match against this rule, a request must satisfy
	// EITHER of the two conditions:
	//
	// - path prefixed with `/foo` AND contains the header `version: v1`
	// - path prefix of `/v2/foo`
	//
	// For example: following match will match all requests.
	//
	//	matches:
	// - path:
	//	   value: "/"
	// +optional
	Matcher *TransformationRequestMatcher `json:"matcher,omitempty"`

	// ClearRouteCache should we clear the route cache if a transformation was matched.
	// +optional
	ClearRouteCache *bool `json:"clearRouteCache,omitempty"`

	// Transformation to apply on the request.
	// +required
	Transformation Transformation `json:"transformation"`
}

// ResponseMatch configures transformations to apply on the response.
type ResponseMatcher struct {
	// Specifies a set of headers that the route should match on. The router will
	// check the response headers against all the specified headers in the route
	// config. A match will happen if all the headers in the route are present in
	// the request with the same values (or based on presence if the value field
	// is not in the config).
	// +optional
	// +kubebuilder:validation:MaxItems=32
	Headers []TransformationHeaderMatcher `json:"matchers,omitempty"`

	// Only match responses with non-empty response code details (this usually
	// implies a local reply).
	// +optional
	ResponseCodeDetails *string `json:"responseCodeDetails,omitempty"`

	// Transformation to apply on the response.
	// +required
	Transformation Transformation `json:"transformation"`
}

// BodyTransformation defines how to transform the body.
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Body' ? has(self.body) : true",message="body must be set when type is Body"
// +kubebuilder:validation:XValidation:rule="self.type == 'MergeJsonKeys' ? has(self.mergeJsonKeys) : true",message="mergeJsonKeys must be set when type is MergeJsonKeys"
// +kubebuilder:validation:AtMostOneOf=body;mergeJsonKeys
type BodyTransformation struct {
	// Type specifies the type of body transformation to apply.
	// +required
	Type BodyTransformationType `json:"type"`

	// Body is the request/response body to be transformed. Only use when Type is Body.
	// +optional
	Body *InjaTemplate `json:"body,omitempty"`

	// MergeJsonKeys is a transformation template used to merge json keys. Only use when Type is MergeJsonKeys.
	// A set of key-value pairs to merge into the JSON body.
	// Each value will be rendered separately, and then placed into the JSON body at
	// the specified key.
	// There are a number of important caveats to using this feature:
	// * This can only be used when the body is parsed as JSON.
	// * This option does NOT work with advanced templates currently
	//
	// Map of key name -> template to render into the JSON body.
	// Specified keys which don't exist in the JSON body will be set,
	// keys which do exist will be override.
	//
	// For example, given the following JSON body:
	// {
	// "key1": "value1"
	// }
	// and the following MergeJsonKeys:
	// {
	// "key1": "{{ header("header1") }}",
	// "key2": "{{ header("header2") }}"
	// }
	// The resulting JSON body will be:
	// {
	// "key1": "header1_value",
	// "key2": "header2_value"
	// }
	// +optional
	MergeJsonKeys map[string]OverridableTemplate `json:"mergeJsonKeys,omitempty"`
}

// BodyTransformationType defines the type of body transformation to apply.
// +kubebuilder:validation:Enum=Body;Passthrough;MergeExtractorsToBody;MergeJsonKeys
type BodyTransformationType string

const (
	// BodyTransformationTypeBody indicates a template-based body transformation
	BodyTransformationTypeBody BodyTransformationType = "Body"
	// BodyTransformationTypePassthrough indicates a passthrough body transformation
	BodyTransformationTypePassthrough BodyTransformationType = "Passthrough"
	// BodyTransformationTypeMergeExtractorsToBody indicates merging extractors to body
	BodyTransformationTypeMergeExtractorsToBody BodyTransformationType = "MergeExtractorsToBody"
	// BodyTransformationTypeMergeJsonKeys indicates merging JSON keys
	BodyTransformationTypeMergeJsonKeys BodyTransformationType = "MergeJsonKeys"
)

// AWSLambdaTransformFormat defines the format used to transform requests/responses
// to/from AWS Lambda functions.
// +kubebuilder:validation:Enum=APIGateway
type AWSLambdaTransformFormat string

const (
	// AWSLambdaFormatAPIGateway transforms the request/response to/from AWS Lambda functions
	// as if it were handled by the AWS API Gateway.
	AWSLambdaFormatAPIGateway AWSLambdaTransformFormat = "APIGateway"
)

// AWSLambdaTransformation defines the AWS Lambda transformation configuration for requests and responses.
type AWSLambdaTransformation struct {
	// RequestFormat defines the format to transform requests to AWS Lambda functions.
	// +optional
	RequestFormat *AWSLambdaTransformFormat `json:"requestFormat,omitempty"`

	// ResponseFormat defines the format to transform responses from AWS Lambda functions.
	// +optional
	ResponseFormat *AWSLambdaTransformFormat `json:"responseFormat,omitempty"`
}
