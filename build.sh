#!/bin/bash

BUILD_VERSION=$(git describe --tags)
BUILD_DATE=$(date -u '+%Y/%m/%d-%H:%M:%S')

echo building $BUILD_VERSION $BUILD_DATE
gox -osarch="darwin/amd64" -osarch="linux/amd64" -osarch="windows/amd64" -ldflags "-X main.version=$BUILD_VERSION -X main.buildDate=$BUILD_DATE" -output "dist/ncd_{{.OS}}_{{.Arch}}"