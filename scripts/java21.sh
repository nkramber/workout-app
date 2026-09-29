#!/usr/bin/env bash
# Print the bin folder of the first Java 21 or later. The Firestore
# emulator of firebase-tools 15 needs Java 21. The script reads JAVA_HOME,
# the Homebrew keg openjdk@21, and the java of PATH, in that order. It
# exits 2 when it finds none.
set -euo pipefail

keg=/opt/homebrew/opt/openjdk@21
from_path=$(dirname "$(command -v java 2>/dev/null || echo /none/java)")

for bin in "${JAVA_HOME:+$JAVA_HOME/bin}" "$keg/bin" "$from_path"; do
  [ -n "$bin" ] && [ -x "$bin/java" ] || continue
  major=$("$bin/java" -version 2>&1 | head -1 | sed -E 's/.*version "([0-9]+).*/\1/')
  if [ "${major:-0}" -ge 21 ] 2>/dev/null; then
    echo "$bin"
    exit 0
  fi
done

echo "needs Java 21 or later. Set JAVA_HOME to a Java 21 JDK." >&2
exit 2
