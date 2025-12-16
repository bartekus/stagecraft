#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
#
# Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
#
# Copyright (C) 2025  Bartek Kus
#
# This program is free software licensed under the terms of the GNU AGPL v3 or later.
#
# See https://www.gnu.org/licenses/ for license details.
#
# run.sh - Runs all checks that match CI workflow
#
# Usage:
#   ./scripts/run.sh

set -euo pipefail

# Deterministic runner for Stagecraft checks.
# Usage:
#   ./scripts/run.sh                # same as: all
#   ./scripts/run.sh all
#   ./scripts/run.sh list
#   ./scripts/run.sh resume
#   ./scripts/run.sh <skill>

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

STATE_DIR=".stagecraft/run"
SKILL_RESULTS_DIR="$STATE_DIR/skills"
LAST_RUN_FILE="$STATE_DIR/last-run.json"
mkdir -p "$SKILL_RESULTS_DIR"

# Colors (deterministic, but still readable)
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1" >&2; }

usage() {
  echo "Usage: $0 [all|list|resume|<skill>]"
  echo ""
  echo "State:"
  echo "  $LAST_RUN_FILE"
  echo "  $SKILL_RESULTS_DIR/<skill>.json"
}

write_skill_result() {
  local skill="$1"
  local status="$2"      # pass|fail|skip
  local exit_code="$3"   # numeric
  local note="${4:-}"
  local out_path="$SKILL_RESULTS_DIR/${skill}.json"

  SKILL="$skill" STATUS="$status" EXIT_CODE="$exit_code" NOTE="$note" OUT_PATH="$out_path" \
  python3 - << 'PY'
import json, os
skill = os.environ["SKILL"]
status = os.environ["STATUS"]
exit_code = int(os.environ["EXIT_CODE"])
note = os.environ.get("NOTE","")
out_path = os.environ["OUT_PATH"]
obj = {"skill": skill, "status": status, "exit_code": exit_code}
if note:
  obj["note"] = note
os.makedirs(os.path.dirname(out_path), exist_ok=True)
with open(out_path, "w", encoding="utf-8") as f:
  json.dump(obj, f, indent=2, sort_keys=True)
  f.write("\n")
PY
}

write_last_run() {
  local overall_status="$1"   # pass|fail
  local failed_csv="$2"       # comma-separated
  local skills_order="$3"     # newline-delimited

  SKILLS_ORDER="$skills_order" OVERALL_STATUS="$overall_status" FAILED_CSV="$failed_csv" OUT_PATH="$LAST_RUN_FILE" \
  python3 - << 'PY'
import json, os
skills_order = [s for s in os.environ.get("SKILLS_ORDER","").split("\n") if s]
overall = os.environ.get("OVERALL_STATUS","fail")
failed = [s for s in os.environ.get("FAILED_CSV","").split(",") if s]
out_path = os.environ["OUT_PATH"]
obj = {"status": overall, "skills": skills_order, "failed": failed}
os.makedirs(os.path.dirname(out_path), exist_ok=True)
with open(out_path, "w", encoding="utf-8") as f:
  json.dump(obj, f, indent=2, sort_keys=True)
  f.write("\n")
PY
}

# Ordered list = deterministic execution order
SKILLS=(
  "lint:gofumpt"
  "lint:golangci"
  "test:build"
  "test:binary"
  "test:go"
  "test:coverage"
  "docs:yaml"
  "docs:validate-spec"
  "docs:spec-reference-check"
  "docs:orphan-docs"
  "docs:orphan-specs"
  "docs:doc-patterns"
  "docs:provider-governance"
  "docs:required-tests"
  "docs:header-comments"
  "docs:spec-sync"
  "docs:feature-integrity"
)

skill_exists() {
  local target="$1"
  for s in "${SKILLS[@]}"; do
    if [ "$s" = "$target" ]; then return 0; fi
  done
  return 1
}

run_skill() {
  local skill="$1"
  local script="scripts/skills/${skill}.sh"

  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "SKILL: $skill"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  if [ ! -x "$script" ]; then
    error "Missing or non-executable: $script"
    write_skill_result "$skill" "fail" 2 "skill script missing or not executable"
    return 2
  fi

  set +e
  "$script"
  local rc=$?
  set -e

  if [ $rc -eq 0 ]; then
    write_skill_result "$skill" "pass" 0
    info "PASS: $skill"
  else
    write_skill_result "$skill" "fail" "$rc"
    error "FAIL: $skill (exit $rc)"
  fi

  return $rc
}

cmd="${1:-all}"

case "$cmd" in
  list)
    for s in "${SKILLS[@]}"; do echo "$s"; done
    exit 0
    ;;

  resume)
    if [ ! -f "$LAST_RUN_FILE" ]; then
      error "No last run file found at $LAST_RUN_FILE"
      echo "Run: $0 all"
      exit 1
    fi

    failed_list="$(python3 - << 'PY'
import json
with open(".stagecraft/run/last-run.json","r",encoding="utf-8") as f:
  obj=json.load(f)
print("\n".join(obj.get("failed",[])))
PY
)"

    if [ -z "$failed_list" ]; then
      info "No failed skills to resume"
      exit 0
    fi

    overall_rc=0
    skills_order=""
    failed_again=()

    while IFS= read -r skill; do
      [ -z "$skill" ] && continue
      skills_order="${skills_order}${skill}\n"
      run_skill "$skill" || { overall_rc=1; failed_again+=("$skill"); }
    done <<EOF
$failed_list
EOF

    failed_csv=$(IFS=,; printf '%s' "${failed_again[*]}")

    if [ $overall_rc -eq 0 ]; then
      write_last_run "pass" "" "$(printf "%b" "$skills_order")"
      info "Resume run passed"
      exit 0
    else
      write_last_run "fail" "$failed_csv" "$(printf "%b" "$skills_order")"
      error "Resume run failed"
      echo "Failed skills: $failed_csv"
      exit 1
    fi
    ;;

  all)
    overall_rc=0
    failed=()
    skills_order=""

    for skill in "${SKILLS[@]}"; do
      skills_order="${skills_order}${skill}\n"
      run_skill "$skill" || { overall_rc=1; failed+=("$skill"); }
    done

    failed_csv=$(IFS=,; printf '%s' "${failed[*]}")

    if [ $overall_rc -eq 0 ]; then
      write_last_run "pass" "" "$(printf "%b" "$skills_order")"
      info "All skills passed"
      exit 0
    else
      write_last_run "fail" "$failed_csv" "$(printf "%b" "$skills_order")"
      error "One or more skills failed"
      echo "Failed skills: $failed_csv"
      echo "Re-run only failures: $0 resume"
      exit 1
    fi
    ;;

  -h|--help|help)
    usage
    exit 0
    ;;

  *)
    if ! skill_exists "$cmd"; then
      error "Unknown command or skill: $cmd"
      usage
      exit 2
    fi

    skills_order="${cmd}\n"
    if run_skill "$cmd"; then
      write_last_run "pass" "" "$(printf "%b" "$skills_order")"
      exit 0
    else
      write_last_run "fail" "$cmd" "$(printf "%b" "$skills_order")"
      exit 1
    fi
    ;;
esac
