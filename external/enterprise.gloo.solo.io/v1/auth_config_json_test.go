package v1

import (
	"encoding/json"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
)

func TestAuthConfigSpecJSONWithSecretReferenceFields(t *testing.T) {
	spec := &AuthConfigSpec{
		Configs: []*AuthConfigSpec_Config{
			{
				AuthConfig: &AuthConfigSpec_Config_OpaAuth{
					OpaAuth: &OpaAuth{
						Modules: []*corev1.SecretReference{
							{Name: "opa-module", Namespace: "default"},
						},
						Query: "data.example.allow",
					},
				},
			},
		},
	}

	marshaled, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTripped AuthConfigSpec
	if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	got := roundTripped.GetConfigs()[0].GetOpaAuth().GetModules()[0]
	if got.Name != "opa-module" || got.Namespace != "default" {
		t.Fatalf("unexpected module ref after round trip: %#v", got)
	}
}

func TestAuthConfigSpecDeepCopyWithSecretReferenceFields(t *testing.T) {
	spec := &AuthConfigSpec{
		Configs: []*AuthConfigSpec_Config{
			{
				AuthConfig: &AuthConfigSpec_Config_OpaAuth{
					OpaAuth: &OpaAuth{
						Modules: []*corev1.SecretReference{
							{Name: "opa-module", Namespace: "default"},
						},
						Query: "data.example.allow",
					},
				},
			},
		},
	}

	var copied AuthConfigSpec
	spec.DeepCopyInto(&copied)

	spec.GetConfigs()[0].GetOpaAuth().GetModules()[0].Name = "mutated"
	got := copied.GetConfigs()[0].GetOpaAuth().GetModules()[0]
	if got.Name != "opa-module" || got.Namespace != "default" {
		t.Fatalf("unexpected module ref after deepcopy: %#v", got)
	}
}

// TestAuthConfigSpecJSONGoStyleDuration is a regression test for issue #1995.
// Existing AuthConfig resources express google.protobuf.Duration fields with
// Go-style strings such as "50ms"/"10s" (the historical jsonpb behavior, which
// parses durations with time.ParseDuration). The k8s 1.35+ JSON bridge must
// keep accepting them; protojson only accepts the canonical seconds form
// ("0.050s") and rejects "50ms", which previously broke AuthConfig apply.
func TestAuthConfigSpecJSONGoStyleDuration(t *testing.T) {
	const raw = `{
	  "configs": [
	    {
	      "name": "passthrough",
	      "passThroughAuth": {
	        "http": {
	          "url": "http://httpbin.default.svc.cluster.local:8000/delay/1",
	          "connectionTimeout": "10s",
	          "responseHeaderTimeout": "50ms"
	        }
	      }
	    }
	  ]
	}`

	var spec AuthConfigSpec
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	http := spec.GetConfigs()[0].GetPassThroughAuth().GetHttp()
	if got := http.GetResponseHeaderTimeout().AsDuration(); got != 50*time.Millisecond {
		t.Fatalf("responseHeaderTimeout = %v, want 50ms", got)
	}
	if got := http.GetConnectionTimeout().AsDuration(); got != 10*time.Second {
		t.Fatalf("connectionTimeout = %v, want 10s", got)
	}
}
