package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"istio.io/istio/pkg/config/crd"
)

// NewPortalValidator creates a CRD validator for portal resources.
func NewPortalValidator(t *testing.T) *crd.Validator {
	t.Helper()

	root := findModuleRoot(t)
	crdDir := filepath.Join(root, "install/generated/portal-crds/templates")
	entries, err := os.ReadDir(crdDir)
	if err != nil {
		t.Fatalf("failed to read CRD directory %s: %v", crdDir, err)
	}

	crdFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".yaml") {
			crdFiles = append(crdFiles, filepath.Join(crdDir, entry.Name()))
		}
	}
	if len(crdFiles) == 0 {
		t.Fatalf("no CRD files found in %s", crdDir)
	}

	v, err := crd.NewValidatorFromFiles(crdFiles...)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	return v
}

// findModuleRoot returns the project root directory by locating go.mod.
func findModuleRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod in any parent directory")
		}
		dir = parent
	}
}
