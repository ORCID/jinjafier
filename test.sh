#!/usr/bin/env bash

# Compile the jinjafier program
go build -o jinjafier

# Run the jinjafier program with default mode (no camelCase splitting)
./jinjafier example.properties
./jinjafier example.yml

# Check if output files match expected (default mode)
git diff --quiet -- example.properties.j2 example.properties.yml example.properties.env.j2 example.yml.env

if [ $? -ne 0 ]; then
    echo "FATAL: default mode output files have changed"
    exit 1
else
    echo "Default mode: no changes detected (pass)"
fi

# Run with -camel-split flag and verify output
./jinjafier -camel-split example.properties
./jinjafier -camel-split example.yml

# Verify camel-split mode produces expected camelCase splitting
if grep -q "ORG_WIBBLE_TEST_CAMEL_CAPS" example.properties.j2 && \
   grep -q "ORG_WIBBLE_BROKER_URL" example.properties.j2 && \
   grep -q "ORG_WIBBLE_MESSAGE_LISTENER_RETRY_COUNT" example.properties.j2 && \
   grep -q "ORG_WIBBLE_CRON_FORMAT" example.properties.j2; then
    echo "Camel-split mode: output verified (pass)"
else
    echo "FATAL: camel-split mode did not produce expected output"
    exit 1
fi

# Restore default mode output
./jinjafier example.properties
./jinjafier example.yml

# Final check that we're back to clean state
git diff --quiet -- example.properties.j2 example.properties.yml example.properties.env.j2 example.yml.env

if [ $? -ne 0 ]; then
    echo "FATAL: failed to restore default mode output"
    exit 1
else
    echo "All tests passed"
fi
