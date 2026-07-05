#!/usr/bin/env bash
set -euo pipefail

mkdir -p build

go build -o build/hoihoi-en .

./build/hoihoi-en patch clean/hoihoi.bin build/hoihoi-candidate.bin

if ! cmp -s build/hoihoi-reference.bin build/hoihoi-candidate.bin; then
  echo "Output differs from build/hoihoi-reference.bin"
  echo
  echo "Reference:"
  sha256sum build/hoihoi-reference.bin
  echo "Candidate:"
  sha256sum build/hoihoi-candidate.bin
  echo
  echo "First differing byte:"
  cmp -l build/hoihoi-reference.bin build/hoihoi-candidate.bin | head -20
  exit 1
fi

echo "OK: candidate output is byte-for-byte identical"
sha256sum build/hoihoi-candidate.bin
