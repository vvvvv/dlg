.PHONY: test
test:
	@exit_code=0; \
	DLG_NO_WARN=1 go test -tags dlg ./tests || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=ERROR go test -tags dlg ./tests/stacktraceerror || exit_code=1; \
	DLG_NO_WARN=1 DLG_STACKTRACE=ALWAYS go test -tags dlg ./tests/stacktracealways || exit_code=1; \
	./tests/assert.sh || exit_code=1; \
	exit $$exit_code

.PHONY: clean
clean:
	rm -f ./examples/example01/example01.dlg.bin
	rm -f ./examples/example01/example01_linker_flags.dlg.bin
	rm -f ./examples/example02/example02.dlg.bin
	rm -f ./examples/example03/example03.bin
	rm -f ./examples/example03/example03.dlg.bin
	rm -f ./examples/example03/example03.dlg.objdump
	rm -f ./examples/example03/example03.objdump
