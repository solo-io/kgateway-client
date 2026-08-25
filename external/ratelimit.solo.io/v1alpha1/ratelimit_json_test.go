package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// setDescriptorSpec mirrors the RateLimitConfig from issue #2626.
func setDescriptorSpec(unit string) string {
	return fmt.Sprintf(`{
	  "raw": {
	    "rateLimits": [
	      {
	        "setActions": [{"remoteAddress": {}}]
	      }
	    ],
	    "setDescriptors": [
	      {
	        "simpleDescriptors": [{"key": "remote_address"}],
	        "rateLimit": {
	          "requestsPerUnit": 2,
	          "unit": %q
	        }
	      }
	    ]
	  }
	}`, unit)
}

// unitJSON returns the encoded unit so a string value can be distinguished from a number.
func unitJSON(t *testing.T, marshaled []byte) string {
	t.Helper()

	var out struct {
		Raw struct {
			SetDescriptors []struct {
				RateLimit struct {
					Unit json.RawMessage `json:"unit"`
				} `json:"rateLimit"`
			} `json:"setDescriptors"`
		} `json:"raw"`
	}
	if err := json.Unmarshal(marshaled, &out); err != nil {
		t.Fatalf("decode marshaled spec: %v", err)
	}
	if len(out.Raw.SetDescriptors) != 1 {
		t.Fatalf("expected one set descriptor, got %d", len(out.Raw.SetDescriptors))
	}
	return string(out.Raw.SetDescriptors[0].RateLimit.Unit)
}

// TestRateLimitConfigSpecJSONExtendedRateLimitUnits covers issue #2626. Pull request
// #2288 updated the Go enum declarations but not the protobuf descriptor used by jsonpb.
// Test both directions because a missing descriptor value makes marshaling emit a number
// without returning an error.
func TestRateLimitConfigSpecJSONExtendedRateLimitUnits(t *testing.T) {
	tests := []struct {
		name string
		unit RateLimit_Unit
	}{
		{name: "MINUTE", unit: RateLimit_MINUTE},
		{name: "MONTH", unit: RateLimit_MONTH},
		{name: "YEAR", unit: RateLimit_YEAR},
		{name: "WEEK", unit: RateLimit_WEEK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec RateLimitConfigSpec
			if err := spec.UnmarshalJSON([]byte(setDescriptorSpec(tt.name))); err != nil {
				t.Fatalf("unmarshal RateLimitConfigSpec with unit %s: %v", tt.name, err)
			}

			setDescriptors := spec.GetRaw().GetSetDescriptors()
			if len(setDescriptors) != 1 {
				t.Fatalf("expected one set descriptor, got %d", len(setDescriptors))
			}
			if got := setDescriptors[0].GetRateLimit().GetUnit(); got != tt.unit {
				t.Fatalf("unit = %s, want %s", got, tt.name)
			}

			marshaled, err := spec.MarshalJSON()
			if err != nil {
				t.Fatalf("marshal RateLimitConfigSpec with unit %s: %v", tt.name, err)
			}
			if got, want := unitJSON(t, marshaled), fmt.Sprintf("%q", tt.name); got != want {
				t.Fatalf("marshaled unit = %s, want %s", got, want)
			}
		})
	}
}

// TestRateLimitConfigSpecJSONRejectsUnknownRateLimitUnit verifies that enum validation
// remains strict after adding the extended units to the descriptor.
func TestRateLimitConfigSpecJSONRejectsUnknownRateLimitUnit(t *testing.T) {
	var spec RateLimitConfigSpec
	err := spec.UnmarshalJSON([]byte(setDescriptorSpec("FORTNIGHT")))
	if err == nil {
		t.Fatal("expected unmarshal to reject unit FORTNIGHT, got no error")
	}
	if !strings.Contains(err.Error(), "FORTNIGHT") {
		t.Fatalf("expected error to name the rejected unit, got: %v", err)
	}
}
