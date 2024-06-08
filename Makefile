.PHONY: test
# test:
# 	$(MAKE) test-spicedb
test: 
	@echo "Running $* tests..."
	gotestsum \
		--format short-verbose \
		--rerun-fails=5 \
		--packages="./..." \
		--junitfile TEST-unit.xml \
		-- \
		-coverprofile=coverage.out \
		-timeout=30m
