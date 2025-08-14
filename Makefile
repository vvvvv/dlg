
.PHONY: test
test:
	@exit_code=0; \
	DLG_NO_WARN=1 go test -tags dlg ./tests/printf || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=ERROR go test -tags dlg ./tests/stacktraceerror || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=ALWAYS go test -tags dlg ./tests/stacktracealways || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ALWAYS go test -tags dlg ./tests/stacktraceregion || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ERROR go test -tags dlg ./tests/stacktraceregiononerror || exit_code=1; \
	./tests/scripts/assert.sh || exit_code=1; \
	exit $$exit_code


BENCH_COUNT          ?= 10
BENCH_TIME           ?= 200ms
BENCH_OUTPUT_LABEL   ?= new
BENCH_BASELINE_LABEL ?= baseline
GOMAXPROCS           ?=

.PHONY:benchmark
benchmark:
	NAME="$(BENCH_OUTPUT_LABEL)" \
	COUNT="$(BENCH_COUNT)" \
	BENCHTIME="$(BENCH_TIME)" \
	GOMAXPROCS="$(GOMAXPROCS)" \
	./tests/scripts/run_benchmarks.sh

.PHONY:benchmark-compare
benchmark-compare:
	NAME="$(BENCH_OUTPUT_LABEL)" \
	BASELINE="$(BENCH_BASELINE_LABEL)" \
	./tests/scripts/compare_benchmarks.sh

.PHONY: benchmark-baseline
benchmark-baseline:
	NAME="baseline" \
	COUNT="$(BENCH_COUNT)" \
	BENCHTIME="$(BENCH_TIME)" \
	GOMAXPROCS="$(GOMAXPROCS)" \
	./tests/scripts/run_benchmarks.sh


.PHONY: clean
clean:
	rm -f ./examples/example01/example01.dlg.bin
	rm -f ./examples/example01/example01_linker_flags.dlg.bin
	rm -f ./examples/example02/example02.dlg.bin
	rm -f ./examples/example03/example03.bin
	rm -f ./examples/example03/example03.dlg.bin
	rm -f ./examples/example03/example03.dlg.objdump
	rm -f ./examples/example03/example03.objdump
