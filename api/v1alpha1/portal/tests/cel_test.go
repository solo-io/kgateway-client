package tests

import (
	"testing"

	"github.com/kgateway-dev/kgateway/v2/test/crvalidation"
)

const (
	// testdataDir is the directory containing the test data for the tests.
	testdataDir = "testdata"
)

func TestCRDs(t *testing.T) {
	v := NewPortalValidator(t)
	crvalidation.TestCRValidation(t, v, testdataDir)
}
