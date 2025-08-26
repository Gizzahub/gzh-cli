#!/usr/bin/env bash
set -euo pipefail

missing=()
for dir in cmd/*; do
  [ -d "$dir" ] || continue
  if [ ! -f "$dir/AGENTS.md" ]; then
    missing+=("$dir/AGENTS.md")
  fi
done

if [ ${#missing[@]} -ne 0 ]; then
  echo "Missing AGENTS.md files:" >&2
  printf ' - %s\n' "${missing[@]}" >&2
  exit 1
fi

echo "All cmd modules contain AGENTS.md"
