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

typeset project_root benchmark_output_dir current_label baseline_label
project_root="$(pkg_root)"
benchmark_output_dir="${OUT:-${project_root}/tests/benchmark_results/}"

# Labels
current_label="${NAME:-new}"
baseline_label="${BASELINE:-baseline}"

# Output file for diff results
typeset comparison_output_file="${benchmark_output_dir}bench.diff.txt"

typeset significance_level change_threshold
significance_level="${ALPHA:-0.05}"
change_threshold="${THRESHOLD:-5}" # percent

# Benchmark pkgs to execute
typeset -a benchmark_pkgs=(
  "${project_root}/tests/printf/"
  "${project_root}/tests/stacktraceregion/"
  "${project_root}/tests/stacktracealways/"
)

# Clear previous comparison file
truncate -s 0 "${comparison_output_file}"

# Compare each benchmark file
for benchmark in "${benchmark_pkgs[@]}"; do
  typeset benchmark_base_name
  benchmark_base_name="$(basename "${benchmark}")"

  typeset baseline_file="${benchmark_output_dir}${benchmark_base_name}.${baseline_label}.bench.txt"
  typeset current_file="${benchmark_output_dir}${benchmark_base_name}.${current_label}.bench.txt"

  benchstat \
    -alpha="${significance_level}" \
    "${baseline_file}" \
    "${current_file}" |
    tee -a "${comparison_output_file}"
done
