#!/bin/bash
set -uo pipefail

# run tests and convert them to junit for circle
go test -v ./cmd | tee go-test.out
PASS=$?
echo ""

# format to junit for circle
mkdir -p ~/junit
cat go-test.out | go-junit-report | tee ~/junit/go-test.xml

exit ${PASS}