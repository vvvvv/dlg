#!/usr/bin/env bash

rm -f ./example01.dlg.bin
go build -tags dlg -o example01.dlg.bin
./example01.dlg.bin
