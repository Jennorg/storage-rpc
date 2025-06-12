#!/bin/bash

# Termina el script si algún comando falla
set -e

echo "Running Makefile to build server and client..."
make all

echo "Build completed. Now running the test script..."
# Ejecuta tu script de pruebas
./scripts/test.sh

echo "All tests have finished."
