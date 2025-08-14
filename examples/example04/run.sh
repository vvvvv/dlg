#!/usr/bin/env bash

rm -f ./example04.dlg.bin
go build -tags dlg -o example04.dlg.bin

printf '==== STARTING WITH DLG_STACKTRACE=REGION,ALWAYS ====\n\n'
DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ALWAYS ./example04.dlg.bin
printf '\n\n'

printf '==== STARTING WITH DLG_STACKTRACE=REGION,ERROR ====\n\n'
DLG_NO_WARN=1 DLG_STACKTRACE=REGION,ERROR ./example04.dlg.bin
