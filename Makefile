.PHONY: validate-refs validate-examples validate-examples-e2e

validate-refs:
ifeq ($(strip $(REFS)),)
	./hack/test-ref-matrix.sh
else
	./hack/test-ref-matrix.sh $(REFS)
endif

validate-examples:
ifeq ($(strip $(REFS)),)
	./hack/test-example-matrix.sh
else
	./hack/test-example-matrix.sh $(REFS)
endif

validate-examples-e2e:
ifeq ($(strip $(REFS)),)
	./hack/test-example-e2e-matrix.sh
else
	./hack/test-example-e2e-matrix.sh $(REFS)
endif
