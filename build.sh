#!/usr/bin/env bash

# Get the latest tag
latestTag=$(git describe --tags)

echo "latest tag: $latestTag"

# Build the Go program, injecting the latest tag into the version variable
echo go build -ldflags "-X main.version=$latestTag" -o jinjafier
go build -ldflags "-X main.version=$latestTag" -o jinjafier

