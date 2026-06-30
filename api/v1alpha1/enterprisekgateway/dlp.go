package enterprisekgateway

// Target controls where DLP masking is applied.
// +kubebuilder:validation:Enum=ResponseBody;AccessLogs
type Target string

const (
	TargetResponseBody Target = "ResponseBody"
	TargetAccessLogs   Target = "AccessLogs"
)

// PresetType identifies a built-in DLP detector bundle.
// The following presets map to subgroup 1 of the listed regex patterns:
//
// SSN:
// - '(?:^|\D)([0-9]{9})(?:\D|$)'
// - '(?:^|\D)([0-9]{3}\-[0-9]{2}\-[0-9]{4})(?:\D|$)'
// - '(?:^|\D)([0-9]{3}\ [0-9]{2}\ [0-9]{4})(?:\D|$)'
//
// Mastercard:
// - '(?:^|\D)(5[1-5][0-9]{2}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4})(?:\D|$)'
//
// Visa:
// - '(?:^|\D)(4[0-9]{3}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4})(?:\D|$)'
//
// Amex:
// - '(?:^|\D)((?:34|37)[0-9]{2}(?:\ |\-|)[0-9]{6}(?:\ |\-|)[0-9]{5})(?:\D|$)'
//
// Discover:
// - '(?:^|\D)(6011(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4})(?:\D|$)'
//
// JCB:
// - '(?:^|\D)(3[0-9]{3}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4}(?:\ |\-|)[0-9]{4})(?:\D|$)'
// - '(?:^|\D)((?:2131|1800)[0-9]{11})(?:\D|$)'
//
// DinersClub:
// - '(?:^|\D)(30[0-5][0-9](?:\ |\-|)[0-9]{6}(?:\ |\-|)[0-9]{4})(?:\D|$)'
// - '(?:^|\D)((?:36|38)[0-9]{2}(?:\ |\-|)[0-9]{6}(?:\ |\-|)[0-9]{4})(?:\D|$)'
//
// CreditCardTrackers:
// - '([1-9][0-9]{2}\-[0-9]{2}\-[0-9]{4}\^\d)'
// - '(?:^|\D)(\%?[Bb]\d{13,19}\^[\-\/\.\w\s]{2,26}\^[0-9][0-9][01][0-9][0-9]{3})'
// - '(?:^|\D)(\;\d{13,19}\=(?:\d{3}|)(?:\d{4}|\=))'
//
// AllCreditCards:
//   - Expands to one action per card type above (Mastercard, Visa, Amex, Discover,
//     JCB, DinersClub, CreditCardTrackers), each with subgroup 1 on every regex
//
// +kubebuilder:validation:Enum=SSN;Mastercard;Visa;Amex;Discover;JCB;DinersClub;CreditCardTrackers;AllCreditCards
type PresetType string

const (
	PresetTypeSSN                PresetType = "SSN"
	PresetTypeMastercard         PresetType = "Mastercard"
	PresetTypeVisa               PresetType = "Visa"
	PresetTypeAmex               PresetType = "Amex"
	PresetTypeDiscover           PresetType = "Discover"
	PresetTypeJCB                PresetType = "JCB"
	PresetTypeDinersClub         PresetType = "DinersClub"
	PresetTypeCreditCardTrackers PresetType = "CreditCardTrackers"
	PresetTypeAllCreditCards     PresetType = "AllCreditCards"
)

const (
	// DefaultDlpMaskChar is the default mask character for custom and keyValue actions.
	// Keep in sync with the +kubebuilder:default markers on CustomAction and KeyValueAction.
	DefaultDlpMaskChar = "X"

	// DefaultDlpPercent is the default percent of the matched string to mask.
	// Keep in sync with the +kubebuilder:default markers on CustomAction and KeyValueAction.
	DefaultDlpPercent int32 = 75
)

// EntDLP configures Data Loss Prevention masking on responses and/or access logs.
type EntDLP struct {
	// Actions to apply, in order.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=32
	// +required
	Actions []DLPAction `json:"actions"`

	// Targets specifies where masking is applied. At least one target must be specified.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:listType=set
	// +required
	Targets []Target `json:"targets"`
}

// DLPAction defines a single masking action.
// +kubebuilder:validation:ExactlyOneOf=preset;custom;keyValue
type DLPAction struct {
	// Preset selects a built-in masking bundle.
	// +optional
	Preset *PresetType `json:"preset,omitempty"`

	// Custom defines a user-provided regex masking action.
	// +optional
	Custom *CustomAction `json:"custom,omitempty"`

	// KeyValue masks a named header or dynamic metadata value.
	// Only affects access logs and response headers, not response bodies.
	// +optional
	KeyValue *KeyValueAction `json:"keyValue,omitempty"`

	// Shadow logs matched patterns without masking. Defaults to false.
	// +optional
	// +kubebuilder:default=false
	Shadow *bool `json:"shadow,omitempty"`
}

// CustomAction defines a user-provided regex masking action.
type CustomAction struct {
	// RegexActions defines the patterns to match and mask.
	// +kubebuilder:validation:MinItems=1
	// +required
	RegexActions []RegexAction `json:"regexActions"`

	// MaskChar is the replacement character. Defaults to X.
	// +optional
	// +kubebuilder:default=X
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1
	MaskChar *string `json:"maskChar,omitempty"`

	// Percent of the matched string to mask. Defaults to 75.
	// +optional
	// +kubebuilder:default=75
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percent *int32 `json:"percent,omitempty"`
}

// KeyValueAction masks the value of a specific header or dynamic metadata key.
type KeyValueAction struct {
	// KeyToMask is the header or metadata key whose value should be masked.
	// +kubebuilder:validation:MinLength=1
	// +required
	KeyToMask string `json:"keyToMask"`

	// MaskChar is the replacement character. Defaults to X.
	// +optional
	// +kubebuilder:default=X
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1
	MaskChar *string `json:"maskChar,omitempty"`

	// Percent of the matched string to mask. Defaults to 75.
	// +optional
	// +kubebuilder:default=75
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Percent *int32 `json:"percent,omitempty"`
}

// RegexAction defines a regex pattern and optional capture group for masking.
type RegexAction struct {
	// An empty regex matches everything, so a minimum length of 1 is enforced.
	// +kubebuilder:validation:MinLength=1
	// +required
	Regex string `json:"regex"`

	// Subgroup selects a capture group to mask. Defaults to 0 (full match).
	// +optional
	// +kubebuilder:validation:Minimum=0
	Subgroup *int32 `json:"subgroup,omitempty"`
}
