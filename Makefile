SHELL=/bin/bash -e -o pipefail
.DEFAULT_GOAL=help
ARGS=$(filter-out $@,$(MAKECMDGOALS))

SOURCE_FILES ?= ./...
TEST_PATTERN ?= .
TEST_OPTIONS ?=
BUILD_DIR=.build
REPORT_DIR=.report
BENCH_PATTERN ?= .

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

bench:
	@go test -bench=$(BENCH_PATTERN) \
 		-run=^$ \
		-benchmem \
		-benchtime=1s \
		-count=2 \
		-cpu=1,2,4

bench-cpu:
	@go test -bench=$(BENCH_PATTERN) -cpuprofile=.report/cpu.prof
	@go tool pprof -http=:8080 .report/cpu.prof

bench-mem:
	@go test -bench=$(BENCH_PATTERN) -memprofile=.report/mem.prof
	@go tool pprof -http=:8080 .report/mem.prof

bench_stat:
	@go install golang.org/x/perf/cmd/benchstat@latest
	@go test -bench=. -benchmem -count=10 > .report/old.txt
	@go test -bench=. -benchmem -count=10 > .report/new.txt
	@benchstat old.txt new.txt

lint: ## Run linter
	@[ -d ${REPORT_DIR} ] || mkdir -p ${REPORT_DIR}
	@golangci-lint run -j 10 --output.code-climate.path ${REPORT_DIR}/qa.json --issues-exit-code 0 || true ; \
 		cat ${REPORT_DIR}/qa.json | jq -r '.[] | "\(.location.path):\(.location.lines.begin) \(.description)"'
