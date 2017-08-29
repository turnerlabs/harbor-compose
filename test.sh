#!/bin/bash
set -uo pipefail

# run tests and convert them to junit for circle
cd $IMPORT_PATH
go test -v ./cmd | tee go-test.out
PASS=$?
echo ""

# format to junit for circle
cat go-test.out | go-junit-report | tee go-test.xml

# move junit test results into special circle directory
sudo mkdir -p ${CIRCLE_TEST_REPORTS}/go-test
sudo mv go-test.xml ${CIRCLE_TEST_REPORTS}/go-test/

exit ${PASS}