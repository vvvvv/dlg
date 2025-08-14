#!/usr/bin/env bash
set -euo pipefail

# Get the root directory of this project
pkg_root() {
  typeset project_root
  project_root="$(realpath -- "$0")"
  project_root="${project_root%/*}/"

  while [[ -n "${project_root}" ]]; do
    typeset go_mod_file="${project_root}/go.mod"
    if [[ -f "${go_mod_file}" ]] && grep --quiet 'module github.com/vvvvv/dlg' "${go_mod_file}" 2>/dev/null; then
      break
    fi
    project_root="${project_root%/*}"
  done
  printf '%s' "${project_root}"
}

typeset project_root benchmark_output_dir benchmark_label
project_root="$(pkg_root)"

# Directory to store benchmark results
benchmark_output_dir="${OUTDIR:-${project_root}/tests/benchmark_results/}"

# Label for the benchmark run (affects output file names)
# Files are named: <benchmark_file>.<benchmark_label>.bench.txt
benchmark_label="${NAME:-new}"

# Benchmark settings
typeset run_count bench_time max_procs
run_count="${COUNT:-10}"
bench_time="${BENCHTIME:-200ms}"
max_procs="${GOMAXPROCS:-}"

if [[ -n "${max_procs}" ]]; then
  export "${max_procs}"
fi

# Benchmark pkgs to execute
typeset -a benchmark_pkgs=(
  "${project_root}/tests/printf/"
  "${project_root}/tests/stacktraceregion/"
  "${project_root}/tests/stacktracealways/"
)

# Run each benchmark file
for benchmark in "${benchmark_pkgs[@]}"; do
  typeset benchmark_base_name
  benchmark_base_name="$(basename "${benchmark}")"

  if [[ -n "${benchmark_label}" ]]; then
    typeset benchmark_result_file="${benchmark_output_dir}${benchmark_base_name}.${benchmark_label}.bench.txt"
  else
    typeset benchmark_result_file="${benchmark_output_dir}${benchmark_base_name}.bench.txt"
  fi

  # Load environment variables specific to this benchmark
  source "${project_root}/tests/scripts/${benchmark_base_name}.env"

  # Run benchmark and save results
  go test \
    -tags dlg \
    -run=^$ \
    -bench=. \
    -benchmem \
    -count="${run_count}" \
    -benchtime="${bench_time}" \
    "${benchmark}" |
    tee "${benchmark_result_file}"
done
