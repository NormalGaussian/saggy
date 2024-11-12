#!/bin/bash

set -euo pipefail

BOLD="\033[1m"
RESET="\033[0m"
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
WHITE="\033[37m"
GREY="\033[90m"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

SAGGY="${SAGGY:-$SCRIPT_DIR/../saggy.sh}"

TMPDIR="$SCRIPT_DIR/tmp"

cleanup() {
    if [ -d "./secrets" ]; then
        rm -rf ./secrets
    fi
    if [ -d "./tmp" ]; then
        rm -rf ./tmp
    fi
}
prepare() {
    mkdir -p ./tmp
}

trap cleanup EXIT

FILTER=${1:-}
if [ -n "$FILTER" ]; then
    # shellcheck disable=SC2010
    # find and xargs require a subshell with lots of exported variables; ls -1 is weak in not handling newlines 
    TESTS=$(find . -type f \( -name "path_*.sh" -o -name "should_*.sh" \) | grep -P "$FILTER")
else
    TESTS=$(find . -type f \( -name "path_*.sh" -o -name "should_*.sh" \))
fi

SUCCESSES=()
FAILURES=()
declare -A FAILURE_CODES
declare -A FAILURE_OUTPUTS
TEST_COUNT=$(echo "$TESTS" | wc -l)
SUCCESS_COUNT=0
SKIPPED_COUNT=0
FAILURE_COUNT=0

# shellcheck disable=SC2016
CHILD_PS4=$GREY'$(printf "%*s" $BASH_SUBSHELL " " | tr " " "+")$(printf "%$((30 - $BASH_SUBSHELL))s" $(basename $0))$(printf "%4s" $LINENO): '$RESET


run_test() {
    local test_script=$1

    if ! [[ "$test_script" =~ $FILTER ]]; then
        echo -e "${BOLD}${WHITE}Skipping $test_script:${RESET}"
        SKIPPED_COUNT=$((SKIPPED_COUNT+1))
        return
    fi

    echo -e "${BOLD}${WHITE}Running $test_script:${RESET}"

    TEST_DIR="$TMPDIR/$test_script"
    TEST_TEMP_DIR="$TEST_DIR/tmp"
    mkdir -p "$TEST_TEMP_DIR"

    if OUTPUT="$(cd "$TEST_DIR" && env -i PS4="$CHILD_PS4" SAGGY="$SAGGY" TMPDIR="$TEST_TEMP_DIR" bash -xeuo pipefail "$SCRIPT_DIR/$test_script" 2>&1)"; then
        SUCCESSES+=("$test_script")
        SUCCESS_COUNT=$((SUCCESS_COUNT+1))
        echo -e "${GREEN}\tSuccess${RESET}"
    else
        FAILURE_CODES[$test_script]=$?
        FAILURES+=("$test_script")
        FAILURE_COUNT=$((FAILURE_COUNT+1))
        FAILURE_OUTPUTS[$test_script]="$OUTPUT"
        echo "$OUTPUT"
        echo -e "${RED}\tFailure - exit code: ${FAILURE_CODES[$test_script]}${RESET}"
        if [[ -n "${STOP_ON_FAILURE:-}" ]]; then
            trap - EXIT
            return 1
        fi
    fi
    cleanup
}

export -f run_test
cleanup
START=$(date +%s%3N)
for test_script in $TESTS; do
    prepare
    if ! run_test "$test_script"; then
        if [[ -n "${STOP_ON_FAILURE:-}" ]]; then
            break
        fi
    fi
done
END=$(date +%s%3N)
ELAPSED=$((END-START))

if [[ "$FAILURE_COUNT" -gt 0 ]]; then
    echo -e "${BOLD}${RED}${FAILURE_COUNT} ${WHITE}Failures:${RESET}"
    for failure in "${FAILURES[@]}"; do
        echo -e "${RED}FAILED${RESET} ${BOLD}${WHITE}$failure${RESET} - exit code: ${FAILURE_CODES[$failure]}"
        echo "${FAILURE_OUTPUTS[$failure]}" | tail -n 3 | sed 's/^/\t/'
    done
fi

echo -e "${BOLD}${WHITE}Ran $TEST_COUNT tests. ${GREEN}$SUCCESS_COUNT${WHITE} succeeded, ${RED}$FAILURE_COUNT${WHITE} failed, ${YELLOW}$SKIPPED_COUNT${WHITE} skipped.${RESET}"
echo "Elapsed time: $ELAPSED milliseconds."

if [ $FAILURE_COUNT -gt 0 ]; then
    exit 1
fi
exit 0
