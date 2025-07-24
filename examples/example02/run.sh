#!/usr/bin/env bash

rm -f example02.dlg.bin
go build -tags dlg -o example02.dlg.bin -ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ERROR'"
./example02.dlg.bin
