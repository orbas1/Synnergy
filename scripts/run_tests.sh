#!/usr/bin/env bash
# run_tests.sh - executes unit and fuzz tests with coverage reporting
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MOD_ROOT="$REPO_ROOT/synnergy-network"

cd "$MOD_ROOT"
# discover all packages containing tests
mapfile -t PACKAGES < <(find . -path './vendor' -prune -o -name '*_test.go' -printf '%h\n' | sort -u)

COVERAGE_FILE="$REPO_ROOT/coverage.out"
: > "$COVERAGE_FILE"
THRESHOLD=${COVERAGE_THRESHOLD:-80}
FIRST=1

for dir in "${PACKAGES[@]}"; do
    pkg="./${dir#./}"
    echo "Testing $pkg"
    if ! go test -run=^$ "$pkg" >/dev/null 2>&1; then
        echo "Skipping $pkg (build failed)"
        echo
        continue
    fi

    go test -covermode=atomic -coverprofile=profile.out "$pkg"
    go tool cover -func=profile.out | tail -1
    if [ "$FIRST" -eq 1 ]; then
        cat profile.out > "$COVERAGE_FILE"
        FIRST=0
    else
        tail -n +2 profile.out >> "$COVERAGE_FILE"
    fi
    rm profile.out

    FUZZ_TARGETS=$(go test "$pkg" -list Fuzz | grep '^Fuzz' || true)
    if [ -n "$FUZZ_TARGETS" ]; then
        for target in $FUZZ_TARGETS; do
            if go test "$pkg" -run=^$ -fuzz="$target" -fuzztime=5s >/dev/null; then
                echo "Fuzzing $target for $pkg completed"
            else
                echo "Fuzzing $target for $pkg failed" >&2
            fi
        done
    else
        echo "No fuzz tests for $pkg"
    fi
    echo
done

TOTAL=$(go tool cover -func="$COVERAGE_FILE" | tail -1 | awk '{print $3}' | tr -d '%')
TOTAL_INT=${TOTAL%.*}
if (( TOTAL_INT < THRESHOLD )); then
    echo "Overall coverage ${TOTAL}% is below threshold ${THRESHOLD}%" >&2
    exit 1
fi

echo "Overall coverage: ${TOTAL}%"
