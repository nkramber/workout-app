#!/usr/bin/env bash
# Prove that the breaking rules of buf.yaml refuse a move of the contract
# to a new package (D-138). The script copies proto/ to proto/moved/ in a
# temporary tree, adds "moved." to each package, and runs buf breaking of
# that tree against this tree. The rules that compare by path pass such a
# move. Only the package rules refuse it. The script exits 1 when buf
# passes the move, or when buf fails for another reason.
# Usage: scripts/package_move_probe.sh <path of buf>
set -euo pipefail

buf=$(cd "$(dirname "$1")" && pwd)/$(basename "$1")
root=$(pwd)
moved=$(mktemp -d)
trap 'rm -rf "$moved"' EXIT

cp buf.yaml "$moved/"
(cd proto && find . -name '*.proto') | while read -r file; do
  mkdir -p "$moved/proto/moved/$(dirname "$file")"
  # An import of the module moves with it. A Google import stays.
  sed -E -e 's/^package ([A-Za-z0-9_.]+);/package moved.\1;/' \
    -e '/^import "google\//! s#^import "#import "moved/#' \
    "proto/$file" > "$moved/proto/moved/$file"
done

if out=$(cd "$moved" && "$buf" breaking --against "$root" 2>&1); then
  echo "package_move_probe: buf breaking passed a move of the contract to a new package."
  echo "Add PACKAGE_NO_DELETE and PACKAGE_SERVICE_NO_DELETE to buf.yaml (D-138)."
  exit 1
fi
if ! grep -q 'Previously present package ".*" was deleted' <<<"$out"; then
  echo "package_move_probe: buf breaking failed for another reason:"
  echo "$out"
  exit 1
fi
echo "package_move_probe: buf breaking refuses a move of the contract to a new package"
