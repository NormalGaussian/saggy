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

TEST_DIR="${TEST_DIR:-$SCRIPT_DIR}"

SAGGY="${SAGGY:-$SCRIPT_DIR/../saggy.sh}"

TMPDIR="$SCRIPT_DIR/tmp"

RESULTS_DIR="$(pwd)/results"
mkdir -p "$RESULTS_DIR"

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
    rm -rf "$RESULTS_DIR"
    mkdir -p "$RESULTS_DIR"
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

TEST_COUNT=$(echo "$TESTS" | wc -l)

# shellcheck disable=SC2016
CHILD_PS4=$GREY'$(printf "%*s" $BASH_SUBSHELL " " | tr " " "+")$(printf "%$((30 - $BASH_SUBSHELL))s" $(basename $0))$(printf "%4s" $LINENO): '$RESET

TESTENV="${TESTENV:-}"
CHILD_ENV=()
IFS="," read -r -a TESTENV_VARS <<< "$TESTENV"
for VAR in "${TESTENV_VARS[@]}"; do
    if [[ ! -z "${!VAR:-}" ]]; then
        CHILD_ENV+=("$VAR=${!VAR}")
    fi
done

CHILD_ENV+=("SAGGY=$SAGGY")
CHILD_ENV+=("PS4=$CHILD_PS4")
CHILD_ENV+=("PATH=$PATH")

run_test() {
    local test_script=$1
    # Avoid collisions as much as possible
    NAME_HASH="$(echo "$test_script" | md5sum - | head -c 4)"
    
    TEST_NAME="${test_script#$TEST_DIR/}"
    TEST_NAME="${TEST_NAME%.sh}"
    TEST_NAME="${TEST_NAME#.}"
    TEST_NAME="${TEST_NAME#/}"
    TEST_NAME="${TEST_NAME//\// }"

    local test_id="$TEST_NAME $NAME_HASH"

    if ! [[ "$test_script" =~ $FILTER ]]; then
        touch "$RESULTS_DIR/$test_id.skipped"
        return
    fi

    CHILD_TEST_DIR="$TMPDIR/$test_script"
    TEST_TEMP_DIR="$CHILD_TEST_DIR/tmp"
    mkdir -p "$TEST_TEMP_DIR"

    cd "$CHILD_TEST_DIR"

    if env -i "${CHILD_ENV[@]}" TMPDIR="$TEST_TEMP_DIR" bash -xeuo pipefail "$SCRIPT_DIR/$test_script" > "$RESULTS_DIR/$test_id.log" 2>&1; then
        touch "$RESULTS_DIR/$test_id.success"
    else
        echo "$?" > "$RESULTS_DIR/$test_id.failure"
        return
    fi
    rm -rf "$CHILD_TEST_DIR"
}

export -f run_test
cleanup
prepare
START=$(date +%s%3N)
for test_script in $TESTS; do
    run_test "$test_script" &
done
wait
END=$(date +%s%3N)
ELAPSED=$((END-START))

SUCCESS_COUNT=0
FAILURE_COUNT=0
SKIPPED_COUNT=0
declare -A FAILURE_CODES
declare -A FAILURE_OUTPUTS
for FILE in "$RESULTS_DIR"/*; do
    if [[ "$FILE" == *.log ]]; then
        continue
    fi
    if [[ "$FILE" == *.skipped ]]; then
        SKIPPED_COUNT=$((SKIPPED_COUNT+1))
        TEST_ID=$(basename "$FILE" .skipped)
        echo -e "${YELLOW}SKIPPED${RESET} ${BOLD}${WHITE}$TEST_ID${RESET}"
        continue
    fi
    if [[ "$FILE" == *.failure ]]; then
        TEST_ID=$(basename "$FILE" .failure)
        FAILURE_COUNT=$((FAILURE_COUNT+1))
        EXIT_CODE=$(cat "$RESULTS_DIR/$TEST_ID.failure")
        OUTPUT=$(tail -n 5 "$RESULTS_DIR/$TEST_ID.log")
        echo -e "${RED}FAILED${RESET} ${BOLD}${WHITE}$TEST_ID${RESET}"
        echo -e "\tExit code: $EXIT_CODE"
        echo "$OUTPUT" | sed 's/^/\t/'
        continue
    fi
    if [[ "$FILE" == *.success ]]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT+1))
        TEST_ID=$(basename "$FILE" .success)
        echo -e "${GREEN}PASSED${RESET} ${BOLD}${WHITE}$TEST_ID${RESET}"
        continue
    fi
    echo "Unknown result file: $FILE"
done

if [[ "$FAILURE_COUNT" -gt 0 ]] && [[ -n "${SAVE_ON_FAILURE:-}" ]]; then
    trap - EXIT
fi

echo -e "${BOLD}${WHITE}Ran $TEST_COUNT tests. ${GREEN}$SUCCESS_COUNT${WHITE} succeeded, ${RED}$FAILURE_COUNT${WHITE} failed, ${YELLOW}$SKIPPED_COUNT${WHITE} skipped.${RESET}"
echo "Elapsed time: $ELAPSED milliseconds."

if [ $FAILURE_COUNT -gt 0 ]; then
    exit 1
fi
exit 0
