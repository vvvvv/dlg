#!/usr/bin/env bash

rm -f ./example01.dlg.bin
go build -tags dlg -o example01.dlg.bin
DLG_STACKTRACE=ERROR ./example01.dlg.bin
