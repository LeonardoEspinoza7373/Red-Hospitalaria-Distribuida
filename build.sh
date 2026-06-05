#!/bin/bash
set -euo pipefail

# Usar Go desde /tmp si está allí, si no, usar el del PATH
if [ -x /tmp/go/bin/go ]; then
    GO=/tmp/go/bin/go
else
    GO=go
fi

BINARY_DIR="dist"
mkdir -p "$BINARY_DIR"

echo "==> Compilando nodo (linux/amd64)..."
GOOS=linux GOARCH=amd64 $GO build -ldflags="-s -w" -o "$BINARY_DIR/nodo" ./cmd/nodo/

echo "==> Compilando proxy (linux/amd64)..."
GOOS=linux GOARCH=amd64 $GO build -ldflags="-s -w" -o "$BINARY_DIR/proxy" ./cmd/proxy/

echo ""
echo "==> Binarios generados:"
ls -lh "$BINARY_DIR/"
