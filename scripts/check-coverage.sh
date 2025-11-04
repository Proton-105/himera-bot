#!/usr/bin/env sh

set -eu

COVERAGE_FILE="coverage.out"
THRESHOLD="70.0"

if [ ! -f "$COVERAGE_FILE" ]; then
  echo "coverage.out not found. Run go test -coverprofile=coverage.out ./... first." >&2
  exit 1
fi

coverage_output=$(go tool cover -func="$COVERAGE_FILE") || {
  echo "Failed to run go tool cover on $COVERAGE_FILE." >&2
  exit 1
}

total=$(printf "%s\n" "$coverage_output" | awk '/^total:/ {gsub("%","",$3); print $3}')

if [ -z "$total" ]; then
  echo "Unable to parse total coverage from $COVERAGE_FILE." >&2
  exit 1
fi

is_below=$(awk -v total="$total" -v threshold="$THRESHOLD" 'BEGIN { if (total+0 < threshold+0) print 1; else print 0 }')

if [ "$is_below" -eq 1 ]; then
  printf "Coverage %s%% is below required %s%%.\n" "$total" "$THRESHOLD"
  exit 1
fi

printf "Coverage %s%% meets the required %s%% threshold.\n" "$total" "$THRESHOLD"
exit 0
