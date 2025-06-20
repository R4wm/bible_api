#!/bin/sh
set -e

# Build the binary
make build

# Install the binary to /usr/local/bin
sudo install -m 755 bible_api /usr/local/bin/bible_api

echo "bible_api has been installed to /usr/local/bin"
