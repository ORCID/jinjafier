#!/usr/bin/env bash

# Compile the jinjafier program
go build -o jinjafier

# Run the jinjafier program with example.properties as input
./jinjafier example.properties
./jinjafier example.yml

# Check if example.properties.j2 and example.properties.yml have changed
git diff --quiet -- example.properties.j2 example.properties.yml example.properties.env.j2 example.yml.env

# If git diff returns a non-zero exit code, the files have changed
if [ $? -ne 0 ]; then
    echo "FATAL: example.properties.j2 or example.properties.yml or example.properties.env.j2 have changed"
    exit 1
else
    echo "No changes in example.properties.j2 or example.properties.yml or example.properties.env.j2"
fi

