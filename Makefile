.DEFAULT_GOAL := help

REL_MAKEFILE_PATH := $(lastword $(MAKEFILE_LIST))
PROJECT_ROOT      := $(abspath $(dir $(REL_MAKEFILE_PATH)))
GO                := go
GO_BIN            := $(shell $(GO) env GOPATH)/bin
BENCHSTAT         := $(GO_BIN)/benchstat

# Tests
GOFLAGS     ?=
TESTFLAGS   ?= -count=1
GO_TEST     := $(GO) test $(GOFLAGS) $(TESTFLAGS)
SCRIPTS_DIR := $(PROJECT_ROOT)/tests/scripts
TESTS_DIR   := $(PROJECT_ROOT)/tests

# Benchmarks
BENCH_COUNT          ?= 10
BENCH_TIME           ?= 200ms
BENCH_OUTPUT_LABEL   ?= new
BENCH_BASELINE_LABEL ?= baseline
GOMAXPROCS           ?=
BENCH_DIR            := $(join $(TESTS_DIR)/, benchmark_results)
COVER_DIR            := $(join $(TESTS_DIR)/, coverage)
COVER_MERGED_DIR     := $(join $(COVER_DIR)/, merged)

# Env Vars for running test suits/code coverage
ENV_printf                  := DLG_NO_WARN=1
ENV_stacktraceerror         := DLG_NO_WARN=1 DLG_STACKTRACE=ERROR
ENV_stacktracealways        := DLG_NO_WARN=1 DLG_STACKTRACE=ALWAYS
ENV_stacktraceregion        := DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ALWAYS
ENV_stacktraceregiononerror := DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ERROR

# Run a test suite and set the correct environment
define run_test
	$2 $(GO_TEST) -tags dlg $(TESTS_DIR)/$1 || exit_code=1;
endef

# Run code coverage and set the correct environment
define run_code_coverage
	mkdir -p $(COVER_DIR)/$1; \
	$2 $(GO) test -tags dlg -count=1 -covermode=atomic -coverpkg ./ ./tests/$1 -args -test.gocoverdir=$(COVER_DIR)/$1 || exit_code=1;
endef

.PHONY: help
help: ## Show this help
	@awk 'BEGIN { FS=":.*## "; print "Targets:" } \
		/^[a-zA-Z0-9_.-]+:.*## / { \
			printf("  %-25s %s\n", $$1, $$2) \
		}' $(REL_MAKEFILE_PATH)

.PHONY: install-deps
install-deps: $(BENCHSTAT) ## Install dependencies

$(BENCHSTAT):
	$(GO) install golang.org/x/perf/cmd/benchstat@latest

#
# Tests
#
.PHONY: test
test: ## Run all tests
	@exit_code=0; \
	$(call run_test,printf,$(ENV_printf)) \
	$(call run_test,stacktraceerror,$(ENV_stacktraceerror)) \
	$(call run_test,stacktracealways,$(ENV_stacktracealways)) \
	$(call run_test,stacktraceregion,$(ENV_stacktraceregion)) \
	$(call run_test,stacktraceregiononerror,$(ENV_stacktraceregiononerror)) \
	$(SCRIPTS_DIR)/assert.sh || exit_code=1; \
	exit $$exit_code

#
# Benchmarks
#
$(BENCH_DIR):
	@mkdir -p $@

.PHONY: benchmark
benchmark: $(BENCH_DIR) ## Run all benchmarks
	NAME="$(BENCH_OUTPUT_LABEL)" \
	COUNT="$(BENCH_COUNT)" \
	BENCHTIME="$(BENCH_TIME)" \
	GOMAXPROCS="$(GOMAXPROCS)" \
	$(SCRIPTS_DIR)/run_benchmarks.sh

.PHONY: benchmark-baseline
benchmark-baseline: $(BENCH_DIR) ## Record a new baseline benchmark set
	BASELINE="$(BENCH_BASELINE_LABEL)" \
	NAME="$(BENCH_OUTPUT_LABEL)" \
	COUNT="$(BENCH_COUNT)" \
	BENCHTIME="$(BENCH_TIME)" \
	GOMAXPROCS="$(GOMAXPROCS)" \
	$(SCRIPTS_DIR)/run_benchmarks.sh

.PHONY: benchmark-compare
benchmark-compare: $(BENCHSTAT) | $(BENCH_DIR) ## Compare benchmarks to baseline benchmarks
	NAME="$(BENCH_OUTPUT_LABEL)" \
	BASELINE="$(BENCH_BASELINE_LABEL)" \
	$(SCRIPTS_DIR)/compare_benchmarks.sh

#
# Code coverage
#
$(COVER_DIR) $(COVER_MERGED_DIR):
	@mkdir -p $@

.PHONY: coverage-run
coverage-run: | $(COVER_DIR) ## Run code coverage
	@exit_code=0; \
	$(call run_code_coverage,printf,$(ENV_printf)) \
	$(call run_code_coverage,stacktraceerror,$(ENV_stacktraceerror)) \
	$(call run_code_coverage,stacktracealways,$(ENV_stacktracealways)) \
	$(call run_code_coverage,stacktraceregion,$(ENV_stacktraceregion)) \
	$(call run_code_coverage,stacktraceregiononerror,$(ENV_stacktraceregiononerror)) \
	exit $$exit_code

.PHONY: coverage-merge
coverage-merge: | $(COVER_MERGED_DIR) ## Merge code coverage and merge into one report
	@$(GO) tool covdata merge \
		-i=$(COVER_DIR)/printf,$(COVER_DIR)/stacktraceerror,$(COVER_DIR)/stacktracealways,$(COVER_DIR)/stacktraceregion,$(COVER_DIR)/stacktraceregiononerror \
		-o=$(COVER_MERGED_DIR)
	@$(GO) tool covdata textfmt -i=$(COVER_MERGED_DIR) -o=$(COVER_DIR)/merged.cover
	@$(GO) tool cover -html=$(COVER_DIR)/merged.cover -o $(COVER_DIR)/coverage.html
	@echo "Coverage report generated: $(COVER_DIR)/coverage.html"
	@$(GO) tool cover -func=$(COVER_DIR)/merged.cover | grep total | awk '{printf ("Total coverage:  %s\n", $$3) }';

.PHONY: coverage
coverage: coverage-run coverage-merge ## Run coverage and merge

#
# Clean
#

.PHONY: clean-coverage
clean-coverage: ## Remove coverage data
	@rm -rf \
	  $(COVER_DIR)/printf \
	  $(COVER_DIR)/stacktraceerror \
	  $(COVER_DIR)/stacktracealways \
	  $(COVER_DIR)/stacktraceregion \
	  $(COVER_DIR)/stacktraceregiononerror \
	  $(COVER_MERGED_DIR) \
	  $(COVER_DIR)/merged.cover \
	  $(COVER_DIR)/coverage.html
	@echo "Cleaned up coverage data"

.PHONY: clean-examples
clean-examples: ## Remove example build artifacts
	@rm -rf \
	  $(PROJECT_ROOT)/examples/example01/example01.dlg.bin \
	  $(PROJECT_ROOT)/examples/example01/example01_linker_flags.dlg.bin \
	  $(PROJECT_ROOT)/examples/example02/example02.dlg.bin \
	  $(PROJECT_ROOT)/examples/example03/example03.bin \
	  $(PROJECT_ROOT)/examples/example03/example03.dlg.bin \
	  $(PROJECT_ROOT)/examples/example03/example03.dlg.objdump \
	  $(PROJECT_ROOT)/examples/example03/example03.objdump
	@echo "Cleaned up examples"

.PHONY: clean-benchmark-results
clean-benchmark-results: ## Remove benchmark results except baseline
	@find '$(BENCH_DIR)' -type f ! -name '*.baseline.bench.txt' -exec rm -f {} +;
	@echo "Cleaned up benchmark results"

.PHONY: clean
clean: clean-examples clean-coverage clean-benchmark-results ## Clean all

