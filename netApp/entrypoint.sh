#!/bin/sh

set -e

echo "Running migrations..."
./build/migrate || { echo "Migration failed"; exit 1; }

echo "Starting the application..."
exec ./build/main