.PHONY: validate-refs

validate-refs:
ifeq ($(strip $(REFS)),)
	./hack/test-ref-matrix.sh
else
	./hack/test-ref-matrix.sh $(REFS)
endif
