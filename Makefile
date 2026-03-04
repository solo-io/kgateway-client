.PHONY: validate-refs validate-examples

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
