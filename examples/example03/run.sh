#!/usr/bin/env bash

rm -f ./example03.bin
rm -f ./example03.dlg.bin
rm -f ./example03.objdump
rm -f ./example03.dlg.objdump

go build -o example03.bin
go build -tags dlg -o example03.dlg.bin

go tool objdump -S example03.bin >example03.objdump
go tool objdump -S example03.dlg.bin >example03.dlg.objdump

printf 'Lines containing "dlg" in disassembly (no build tag): %5d    (see: %s)\n' "$(grep 'dlg' ./example03.objdump | grep -v '^TEXT' | wc -l)" 'example03.objdump'
printf 'Lines containing "dlg" in disassembly (dlg build tag): %5d   (see: %s)\n' "$(grep 'dlg' ./example03.dlg.objdump | grep -v '^TEXT' | wc -l)" 'example03.dlg.objdump'
