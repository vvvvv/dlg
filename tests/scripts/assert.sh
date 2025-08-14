#!/usr/bin/env bash

# set -x

_err(){
  printf '%s ' "$@" >&2
  printf '\n' >&2
  exit 200
}

typeset -i ok_tests=0 failed_tests=0 total_tests=0

_test_header(){
  total_tests=$(( total_tests + 1))
  printf -- ' --- Testing %s\n' "${1}"
}

_test_ok(){
  ok_tests=$(( ok_tests + 1))
  printf -- '       %-8s\n\n' 'OK'
}

_test_failed(){
  failed_tests=$(( failed_tests + 1))
  printf -- '       %-10s%s\n' 'FAILED' "${1}" 
  if [[ ! -z "${2:+x}" ]]; then
    while read -r line; do 
      printf -- '         %s\n' "${line}"

    done <<<"${2}"
  fi
  if [[ ! -z "${3:+x}" ]]; then
    printf -- '       %s\n\n' "${3}" 
  fi
  printf '\n'
}

_test_synopses(){
  printf '\n'
  printf -- '   Ran %s tests\n     %-8s%4s\n     %-8s%4s\n' "${total_tests}" "OK" "${ok_tests}" "FAILED" "${failed_tests}"
  exit "${failed_tests}"
}

typeset old_pwd="${PWD}"

# Get absolute path to this packages root dir.
# Starting from this scripts parent dir we traverse towards root.
# The directory which contains this packages go.mod is the root dir.
typeset pkg_root
pkg_root="$(realpath -- "$0")"
pkg_root="${pkg_root%/*}/"

while [[ -n "${pkg_root}" ]]; do
  typeset go_mod="${pkg_root}/go.mod"
  if [[ -f "${go_mod}" ]] && grep --quiet 'module github.com/vvvvv/dlg' "${go_mod}" 2>/dev/null; then
    break
  fi
  pkg_root="${pkg_root%/*}"
done

if [[ -z "${pkg_root}" ]]; then
  _err "Failed to find go module root directory"
fi

# Version of dlg to test against
# Used for the test codes go.mod file
typeset version="${DLG_VERSION:-$(( cd "${pkg_root}"; git rev-list --all 2>/dev/null || echo "v0.0.0" ) | head -n1 )}"
version="v0.0.0" #TODO: use the git hash

# Temp dir for building the test code
typeset tmp_dir
tmp_dir="$(mktemp -d)"

# This dir is going to be removed on error/exit - better make sure it isn't empty
if [[ "$?" -ne 0 ]]; then
  _err "Failed to create tmp dir"
fi

# Cleanup temp files on error/exit
trap "cd ${old_pwd}; rm -rf ${tmp_dir}" SIGINT SIGTERM EXIT

# cd into tmp_dir to start building the test code
cd "${tmp_dir}"

typeset test_mod_name="assert_no_dlg"

# Create go module 
go mod init "${test_mod_name}" 2>/dev/null

if [[ "$?" -ne 0 || ! -s 'go.mod' ]]; then
  _err "Failed to initialize go module"
fi


# Change go.mod file so the local dlg package is being used
cat >> go.mod <<END

require github.com/vvvvv/dlg ${version}

replace github.com/vvvvv/dlg => ${pkg_root}

END

typeset test_str="hi from fmt.Println"

cat > main.go <<END
package main

import (
  "fmt"
  "os"
  "github.com/vvvvv/dlg"
)

func main(){
  fmt.Println("${test_str}")
  dlg.StartTrace()
  dlg.Printf("message from dlg")
  dlg.StopTrace()
  dlg.SetOutput(os.Stdout)
}

END

typeset bin_name="out"

_test_header "if code compiles without errors"
typeset go_build_out
go_build_out="$(go build -o "${bin_name}" 2>&1)"
if [[ "$?" -ne 0 ]]; then
  _test_failed "${go_build_out}"
else
  _test_ok
fi


_test_header "if test string is being output"
typeset test_output
test_output="$(./${bin_name})"
if [[ "${test_output}" != "${test_str}" ]]; then 
  _test_failed "expected: '${test_str}' ; got: '${test_output}'"
else
  _test_ok 
fi


_test_header "if dlg API is not in compiled output when build without dlg tag"
go tool objdump "${bin_name}" 2>/dev/null 1> objdump
if grep --quiet 'dlg.Printf' 'objdump'; then
# if ! go tool objdump "${bin_name}" | grep --quiet 'main'; then
  _test_failed "expected binary to not contain any reference to dlg.Printf but got:" "$(grep -A2 -B2 'dlg.Printf' 'objdump' )"
else
  _test_ok
fi

# Delete binary to recompile with dlg tag
rm "${bin_name}"

_test_header "if debug banner is being printed when build with dlg"
go_build_out="$(go build -tags dlg -o "${bin_name}" 2>&1)"
test_output="$(./${bin_name} 2>&1)"
if ! grep --quiet 'DEBUG BUILD' <<<"${test_output}"; then 
  _test_failed "expected DEBUG BUILD OUTPUT"
else
  _test_ok
fi

_test_synopses 
