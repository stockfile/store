#!/usr/bin/env bash

echo "Compiling functions to bin/handlers/ ..."

rm -rf bin/

cd src/handlers/
for f in *.go; do
  filename="${f%.go}"
  if GOOS=linux go build -ldflags="-s -w" -o "../../bin/handlers/$filename" ${f}; then
    upx "../../bin/handlers/${filename%.*}"
    echo "✓ Compiled $filename"
  else
    echo "✕ Failed to compile $filename!"
    exit 1
  fi
done

echo "Done."
