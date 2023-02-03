#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------
DEPSGOBIN:=$(shell pwd)/.bin
export PATH:=$(DEPSGOBIN):$(PATH)
export GOBIN:=$(DEPSGOBIN)

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

.PHONY: install-go-tools
install-go-tools:
	go install golang.org/x/tools/cmd/goimports

.PHONY: format-code
format-code: install-go-tools
	goimports -w .

#----------------------------------------------------------------------------------
# Tests
#----------------------------------------------------------------------------------

GINKGO_VERSION ?= 2.5.0 # match our go.mod
GINKGO_ENV ?= GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore ACK_GINKGO_DEPRECATIONS=$(GINKGO_VERSION)
GINKGO_FLAGS ?= -v -tags=purego -compilers=4 -fail-fast -randomize-suites -randomize-all -skip-package=./installutils/kubeinstall,./debugutils/test
GINKGO_REPORT_FLAGS ?= --json-report=test-report.json --junit-report=junit.xml -output-dir=$(OUTPUT_DIR)
GINKGO_COVERAGE_FLAGS := --cover --covermode=count --coverprofile=coverage.cov

# This is a way for a user executing `make test` to be able to provide flags which we do not include by default
# For example, you may want to run tests multiple times, or with various timeouts
GINKGO_USER_FLAGS ?=

.PHONY: install-test-tools
install-test-tools:
	go install github.com/onsi/ginkgo/v2/ginkgo@v$(GINKGO_VERSION)

.PHONY: test
test: install-test-tools ## Run tests in the {TEST_PKG}
	$(GINKGO_ENV) ginkgo \
		$(GINKGO_FLAGS) $(GINKGO_REPORT_FLAGS) $(GINKGO_USER_FLAGS) \
		$(TEST_PKG)

.PHONY: test-with-coverage ## Run tests in the {TEST_PKG} with coverage
test-with-coverage: GINKGO_FLAGS += $(GINKGO_COVERAGE_FLAGS)
test-with-coverage: test
	go tool cover -html coverage.cov