SHELL=/bin/bash -e -o pipefail
.DEFAULT_GOAL=help
ARGS=$(filter-out $@,$(MAKECMDGOALS))

SOURCE_FILES ?= ./...
TEST_PATTERN ?= .
TEST_OPTIONS ?=
BUILD_DIR=.build
REPORT_DIR=.report

help: ## Show help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

define test_reports
	cat ${REPORT_DIR}/unit.txt | go-junit-report > ${REPORT_DIR}/unit.xml
	gocover-cobertura < ${REPORT_DIR}/coverage.txt > ${REPORT_DIR}/coverage.xml
endef

test: ## Run unit tests
	@[ -d ${REPORT_DIR} ] || mkdir -p ${REPORT_DIR}
	@go install github.com/jstemmer/go-junit-report@latest
	@go install github.com/boumenot/gocover-cobertura@latest

	@go test $(TEST_OPTIONS) \
		-failfast \
		-race \
		-coverpkg=./... \
		-covermode=atomic \
		-coverprofile=${REPORT_DIR}/coverage.txt $(SOURCE_FILES) \
		-run $(TEST_PATTERN) \
		-timeout=2m \
		-v | tee ${REPORT_DIR}/unit.txt && \
			$(call test_reports,) || \
			$(call test_reports,); \
			exit $$?

lint: ## Run linter
	@[ -d ${REPORT_DIR} ] || mkdir -p ${REPORT_DIR}
	golangci-lint run --timeout 5m --out-format code-climate | tee ${REPORT_DIR}/qa.json | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'
