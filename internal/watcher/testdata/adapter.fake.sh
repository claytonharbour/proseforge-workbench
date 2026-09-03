#!/usr/bin/env bash
# ADAPTER: fake — proves the SEAM without a credential or a model.
#
# This is not scaffolding: it is the positive control for the adapter contract.
# Because it exists, dispatch, argument order and exit-code classification are all
# testable on a machine with no auth and no model at all.
#
#   argv    $1 batch.json   $2 prompt file   $3 active-work.json ("" if none)
#   stdout  the worker's reply, or the literal NO_POST
#   exit    0 ok · 10 cannot wake · 1 failed
#
# FAKE_REPLY / FAKE_EXIT / FAKE_SLEEP drive it.
set -euo pipefail
BATCH="${1:?}"; PROMPT="${2:?}"; ACTIVE="${3:-}"

# Argument-order arm: the adapter asserts what it was handed, so a caller that
# swaps argv fails loudly here rather than silently waking with a prompt as batch.
[[ -r "$BATCH"  ]] || { echo "adapter got no readable batch: $BATCH" >&2; exit 1; }
[[ -r "$PROMPT" ]] || { echo "adapter got no readable prompt: $PROMPT" >&2; exit 1; }

[[ -n "${FAKE_SLEEP:-}" ]] && sleep "$FAKE_SLEEP"

exit_code="${FAKE_EXIT:-0}"
[[ "$exit_code" != "0" ]] && exit "$exit_code"

printf '%s\n' "${FAKE_REPLY:-NO_POST}"
printf 'saw batch=%s prompt=%s active=%s wake=%s\n' \
  "$(basename "$BATCH")" "$(basename "$PROMPT")" \
  "${ACTIVE:+$(basename "$ACTIVE")}" "${WATCHER_WAKE_ID:-none}" >&2
